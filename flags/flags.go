/**
 * @file flags
 * @brief Parse environment variables and command-line arguments
 *
 * Parse environment variables and command-line arguments and save them into Options
 */

package flags

import (
	// System
	"bytes"
	"fmt"
	"os"
	"strings"

	// Third-party
	"github.com/jessevdk/go-flags"
	// Project
)

/**
 * @class Options
 * @brief Collect all environment variables and command-line arguments
 */
type Options struct {
	//TODO: add all needed options
	LogFile    flags.Filename `long:"logfile" short:"l" description:"FILE"` ///< command-line argument: filename of log file
	ConfigFile flags.Filename `long:"config" short:"c" description:"FILE"`  ///< command-line argument: filename of json config file
	Debug      bool           `long:"debug" short:"d" description:""`       ///< command-line flag: debug if given
}

/**
 * @brief Parse command-line and environment arguments, validate it and return
 * @return opts Pointer to filled Options object
 */
func NewFlags() *Options {
	opts := new(Options)
	p := flags.NewParser(opts, flags.HelpFlag)
	args, err := p.Parse()

	if len(args) > 0 {
		fatal(p, "Get extra params: "+strings.Join(args, " "), true)
	}

	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fatal(p, "Can't parse command-line arguments", false)
			os.Exit(1)
		}
	}

	return opts
}

/**
 * @brief Print message to STDERR, if with_help, print usage message and exit with error code
 * @param[in] p Parser object to write usage help
 * @param[in] message Message string
 * @param[in] with_help Print usage help before exit if given
 */
func fatal(p *flags.Parser, message string, with_help bool) {
	var b bytes.Buffer

	if with_help {
		p.WriteHelp(&b)
	}

	fmt.Fprintln(os.Stderr, message, "\n", b.String())
	os.Exit(1)
}
