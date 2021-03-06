package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
)

// StepHelmApplyOptions contains the command line flags
type StepHelmApplyOptions struct {
	StepHelmOptions

	Namespace   string
	ReleaseName string
	Wait        bool
}

var (
	StepHelmApplyLong = templates.LongDesc(`
		Applies the helm chart in a given directory.

		This step is usually used to apply any GitOps promotion changes into a Staging or Production cluster.
`)

	StepHelmApplyExample = templates.Examples(`
		# apply the chart in the env folder to namespace jx-staging 
		jx step helm apply --dir env --namespace jx-staging

`)
)

func NewCmdStepHelmApply(f Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := StepHelmApplyOptions{
		StepHelmOptions: StepHelmOptions{
			StepOptions: StepOptions{
				CommonOptions: CommonOptions{
					Factory: f,
					Out:     out,
					Err:     errOut,
				},
			},
		},
	}
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Applies the helm chart in a given directory",
		Aliases: []string{""},
		Long:    StepHelmApplyLong,
		Example: StepHelmApplyExample,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			CheckErr(err)
		},
	}
	options.addStepHelmFlags(cmd)

	cmd.Flags().StringVarP(&options.Namespace, "namespace", "", "", "The kubernetes namespace to apply the helm chart to")
	cmd.Flags().StringVarP(&options.ReleaseName, "name", "", "", "The name of the release")
	cmd.Flags().BoolVarP(&options.Wait, "wait", "", true, "Wait for Kubernetes readiness probe to confirm deployment")
	return cmd
}

func (o *StepHelmApplyOptions) Run() error {
	dir := o.Dir
	helmBinary, err := o.helmInitDependencyBuild(dir, o.defaultReleaseCharts())
	if err != nil {
		return err
	}

	ns := o.Namespace
	if ns == "" {
		ns = os.Getenv("DEPLOY_NAMESPACE")
	}
	if ns == "" {
		return fmt.Errorf("No --namespace option specified or $DEPLOY_NAMESPACE environment variable available")
	}

	releaseName := o.ReleaseName
	if releaseName == "" {
		if helmBinary == "helm" {
			releaseName = ns
		} else {
			releaseName = "jx"
		}
	}
	info := util.ColorInfo
	log.Infof("Applying helm chart at %s as release name %s to namespace %s\n", info(dir), info(releaseName), info(ns))

	if o.Wait {
		timeout := 600
		err = o.Helm().UpgradeChart(dir, releaseName, ns, nil, true, &timeout, false, true, nil, nil)
	} else {
		err = o.Helm().UpgradeChart(dir, releaseName, ns, nil, true, nil, false, false, nil, nil)
	}
	if err != nil {
		return err
	}
	return nil
}
