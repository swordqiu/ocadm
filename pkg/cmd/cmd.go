package cmd

import (
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/alpha"
)

func NewOneCloudAdminCommand(in io.Reader, out, err io.Writer) (*cobra.Command, []func() *cobra.Command) {
	cmds := &cobra.Command{
		Use:   "ocadm",
		Short: "Deploy and manage onecloud services on kubernetes cluster",
	}

	/*kubeadm := func() *cobra.Command {
		return kubeadmcmd.NewKubeadmCommand(os.Stdin, os.Stdout, os.Stderr)
	}*/

	cmds.ResetFlags()

	cmds.AddCommand(NewCmdConfig(out))
	cmds.AddCommand(NewCmdInit(out, nil))
	cmds.AddCommand(NewCmdJoin(out, nil))
	cmds.AddCommand(NewCmdReset(in, out, nil))
	cmds.AddCommand(NewCmdToken(out, err))
	cmds.AddCommand(alpha.NewCmdAlpha(in, out))

	cmds.AddCommand(NewCmdCluster(out))
	cmds.AddCommand(NewCmdComponent(out))
	cmds.AddCommand(NewCmdNode(out))
	cmds.AddCommand(NewCmdBaremetal(out))
	cmds.AddCommand(NewCmdVersion(out))
	cmds.AddCommand(NewCmdLonghorn(out))

	commandFns := []func() *cobra.Command{}

	for i := range commandFns {
		cmds.AddCommand(commandFns[i]())
	}

	return cmds, commandFns
}
