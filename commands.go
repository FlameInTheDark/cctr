package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"time"

	"github.com/cluttrdev/deepl-go/deepl"
	"github.com/urfave/cli/v3"
)

func translate() *cli.Command {
	return &cli.Command{
		Name:  "translate",
		Usage: "Translate half-life closed captions file to specified language",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "deepl api key",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "Language code to translate to",
				Value:   "EN",
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage:   "output file",
			},
		},
		ArgsUsage: "<path>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "path",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			key := c.String("key")
			if key == "" {
				return errors.New("--key must be specified")
			}
			path := c.StringArg("path")
			if path == "" {
				return errors.New("<path> must be specified")
			}
			outputName := c.String("out")
			if outputName == "" {
				outputName = fmt.Sprintf("out_%d.txt", time.Now().UnixNano())
			}

			translator, err := deepl.NewTranslator(key)
			if err != nil {
				return err
			}

			usage, err := translator.GetUsage()
			if err != nil {
				return err
			}

			var limit = usage.CharacterLimit - usage.CharacterCount

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			var localeContent string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				localeContent += scanner.Text() + "\n"
			}

			locale, err := ParseLocaleFile(localeContent)
			if err != nil {
				return err
			}
			symbols := locale.SymbolCount()
			fmt.Println("Total symbols to translate: ", symbols)

			if limit < symbols {
				return errors.New("not enough free characters to translate this file, upgrade your deepl plan")
			}

			var toTranslate []string
			for _, token := range locale.Tokens {
				toTranslate = append(toTranslate, token.Value)
			}

			chunks := SplitByMaxBytes(toTranslate, 10000)
			fmt.Println("Text split into chunks: ", len(chunks))

			// Initialize progress bar
			pw := progress.NewWriter()
			pw.SetOutputWriter(os.Stdout)
			pw.SetAutoStop(false)
			pw.SetTrackerLength(25)
			pw.SetMessageWidth(24)
			pw.SetNumTrackersExpected(1)
			pw.SetSortBy(progress.SortByPercentDsc)
			pw.SetStyle(progress.StyleDefault)
			pw.SetTrackerPosition(progress.PositionRight)
			pw.SetUpdateFrequency(time.Millisecond * 100)
			pw.Style().Options.PercentFormat = "%4.1f%%"
			pw.Style().Visibility.ETA = true
			pw.Style().Visibility.ETAOverall = false
			pw.Style().Visibility.Speed = true
			pw.Style().Visibility.SpeedOverall = false
			pw.Style().Visibility.Time = true
			pw.Style().Visibility.TrackerOverall = true
			pw.Style().Visibility.Value = true
			pw.Style().Visibility.Pinned = true

			// Create a tracker for translation progress
			totalChunks := int64(len(chunks))
			translationTracker := progress.Tracker{
				Message: "Translating chunks",
				Total:   totalChunks,
				Units:   progress.UnitsDefault,
			}
			pw.AppendTracker(&translationTracker)

			// Start the progress writer in a separate goroutine
			go pw.Render()

			var skipped int
			// Translate chunks and update progress
			for i, ch := range chunks {
				translations, err := translator.TranslateText(ch, c.String("lang"))
				if err != nil {
					skipped++
					continue
				}
				for it, t := range translations {
					chunks[i][it] = t.Text
				}
				translationTracker.Increment(1)
			}

			if skipped > 0 {
				fmt.Printf("Skipped %d chunks due to API errors\nSaving with partial translation if possible\n", skipped)
			}

			// Mark tracker as complete and stop the progress writer
			translationTracker.MarkAsDone()
			pw.Stop()
			time.Sleep(time.Millisecond * 100) // Give time for the final render

			for i, t := range MergeChunks(chunks) {
				locale.Tokens[i].Value = t
			}

			outf, err := os.OpenFile(outputName, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer outf.Close()

			_, err = outf.WriteString(WriteLocaleFile(locale))
			if err != nil {
				return err
			}
			fmt.Println("Done!")
			return nil
		},
	}
}

func count() *cli.Command {
	return &cli.Command{
		Name:        "count",
		Usage:       "Calculates the amount of symbols to translate",
		Description: "If API key is provided, will show the comparison with the plan limits.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Usage:   "deepl api key",
			},
		},
		ArgsUsage: "<path>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "path",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			path := c.StringArg("path")
			if path == "" {
				return errors.New("<path> must be specified")
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			var localeContent string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				localeContent += scanner.Text() + "\n"
			}

			locale, err := ParseLocaleFile(localeContent)
			if err != nil {
				return err
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"#", "Count"})

			symbols := locale.SymbolCount()
			t.AppendRows([]table.Row{
				{"To translate", symbols},
			})
			t.AppendSeparator()

			key := c.String("key")
			if key != "" {
				translator, err := deepl.NewTranslator(key)
				if err != nil {
					return err
				}

				usage, err := translator.GetUsage()
				if err != nil {
					return err
				}
				t.AppendRows([]table.Row{
					{"Used characters", fmt.Sprintf("%d/%d", usage.CharacterCount, usage.CharacterLimit)},
				})

				left := usage.CharacterCount + symbols
				if left > usage.CharacterLimit {
					left = usage.CharacterLimit
				}
				t.AppendRows([]table.Row{
					{"After translation", fmt.Sprintf("%d/%d", left, usage.CharacterLimit)},
				})

				if usage.CharacterLimit-usage.CharacterCount < symbols {
					t.AppendRows([]table.Row{
						{"Not enough free characters to translate this file, upgrade your deepl plan"},
					}, table.RowConfig{AutoMerge: true})
				}
			}
			t.SetStyle(table.StyleLight)
			t.Render()
			return nil
		},
	}
}

func languages() *cli.Command {
	return &cli.Command{
		Name:        "languages",
		Usage:       "Show list of supported languages",
		Description: "Required API key to work.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Aliases:  []string{"k"},
				Usage:    "deepl api key",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			translator, err := deepl.NewTranslator(c.String("key"))
			if err != nil {
				return err
			}

			lang, err := translator.GetLanguages("target")
			if err != nil {
				return err
			}
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Language", "Code"})
			for _, l := range lang {
				t.AppendRows([]table.Row{
					{l.Name, l.Code},
				})
			}
			t.SetStyle(table.StyleLight)
			t.Render()
			return nil
		},
	}
}
