package cmd

import (
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func docsCommand(dst *cobra.Command) *cobra.Command {
	var format []string

	var docsCmd = &cobra.Command{
		Use:   "docs",
		Short: "Generate linstor-iscsi documentation",
		Run: func(cmd *cobra.Command, args []string) {

			for _, f := range format {
				dir := path.Join("./docs", f)
				_ = os.Mkdir(dir, 0755) // we don't care, if this fails, the next one fails
				switch f {
				case "man":
					header := &doc.GenManHeader{
						Title:   "linstor-iscsi",
						Section: "3",
					}
					if err := doc.GenManTree(dst, header, dir); err != nil {
						log.Fatal(err)
					}
				case "md":
					if err := doc.GenMarkdownTree(dst, dir); err != nil {
						log.Fatal(err)
					}
				}
			}
		},
	}

	docsCmd.ResetCommands()
	docsCmd.Flags().StringSliceVar(&format, "format", []string{"md"}, "Generate documentation in the given format (md,man)")
	docsCmd.DisableAutoGenTag = true

	return docsCmd
}
