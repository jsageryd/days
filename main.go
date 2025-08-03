package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	args := os.Args[1:]

	printUsage := func() {
		fmt.Fprintln(os.Stderr, "Usage: days [<from> <to>|<+-days>]")
	}

	if len(args) == 1 || len(args) > 3 {
		printUsage()
		os.Exit(1)
	}

	from := time.Now()
	to := from

	if len(args) == 2 {
		var err error

		from, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			printUsage()
			os.Exit(1)
		}

		to, err = time.Parse("2006-01-02", args[1])
		if err != nil {
			delta, err := parseDelta(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing date or delta %q\n", args[1])
				printUsage()
				os.Exit(1)
			}

			if delta > 0 {
				to = from.AddDate(0, 0, delta)
			} else {
				to = from
				from = to.AddDate(0, 0, delta)
			}
		}
	}

	if to.Before(from) {
		from, to = to, from
	}

	cStr := calendar(from, to)

	fmt.Println(cStr)

	if !from.Equal(to) {
		workday, weekend := days(from, to)
		days := workday + weekend

		fmt.Println()
		fmt.Printf("%s - %s: ", from.Format("2006-01-02"), to.Format("2006-01-02"))

		pl := func(n int) string {
			if n == 1 {
				return ""
			}
			return "s"
		}

		switch {
		case workday > 0 && weekend > 0:
			fmt.Printf("%d day%s (%d work day%s + %d weekend day%s)", days, pl(days), workday, pl(workday), weekend, pl(weekend))
		case workday > 0 && weekend == 0:
			fmt.Printf("%d work day%s", workday, pl(workday))
		case workday == 0 && weekend > 0:
			fmt.Printf("%d weekend day%s", weekend, pl(weekend))
		}

		fmt.Printf(" (%d night%s)\n", days-1, pl(days-1))
	}
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing date %q: %w", s, err)
	}
	return t, nil
}

func parseDelta(s string) (days int, err error) {
	return strconv.Atoi(s)
}

func days(from, to time.Time) (workday, weekend int) {
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		switch d.Weekday() {
		case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
			workday++
		case time.Saturday, time.Sunday:
			weekend++
		}
	}

	return workday, weekend
}

func calendar(from, to time.Time) string {
	var buf bytes.Buffer

	fmt.Fprint(&buf, fg("     Mon Tue Wed Thu Fri Sat Sun", 231))

	first := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
	last := time.Date(to.Year(), to.Month()+1, 1, 0, 0, 0, 0, to.Location()).AddDate(0, 0, -1)

	var newMonth bool

	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		// New month
		if d.Day() == 1 {
			newMonth = true
			buf.WriteByte('\n')

			fmt.Fprintf(&buf, "%s  ", fg(d.Format("Jan"), 231))

			start := d

			for start.Weekday() != time.Monday {
				start = start.AddDate(0, 0, -1)
			}

			for blank := start.AddDate(0, 0, 1); !blank.After(d); blank = blank.AddDate(0, 0, 1) {
				fmt.Fprint(&buf, "    ")
			}
		} else if d.Weekday() == time.Monday {
			newMonth = false
			fmt.Fprint(&buf, "     ")
		}

		dayFormatStr := "%3d"

		// Highlight today
		if d.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			dayFormatStr = invert(dayFormatStr)
		}

		if !d.Before(from) && !d.After(to) {
			switch d.Weekday() {
			case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
				dayFormatStr = fg(dayFormatStr, 39) // Blue for weekdays
			case time.Saturday, time.Sunday:
				dayFormatStr = fg(dayFormatStr, 197) // Red for weekends
			}
		} else {
			dayFormatStr = fg(dayFormatStr, 251) // Grey for out of range days
		}

		// Print day
		fmt.Fprintf(&buf, dayFormatStr, d.Day())

		// Space between days
		if d.Weekday() != time.Sunday {
			fmt.Fprint(&buf, " ")
		}

		// Print year if first week of span or January
		if d.Weekday() == time.Sunday && newMonth {
			if (d.Year() == first.Year() && d.Month() == first.Month()) || d.Month() == time.January {
				fmt.Fprintf(&buf, "  "+fg("%d", 231), d.Year())
			}
		}

		// Newline after Sunday
		if d.Weekday() == time.Sunday && !d.Equal(last) {
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

func invert(s string) string {
	return "\033[7m" + s + "\033[0m"
}

func fg(s string, c int) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\033[0m", c, s)
}
