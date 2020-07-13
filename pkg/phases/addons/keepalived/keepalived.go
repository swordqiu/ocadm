package keepalived

import (
	// "errors"
	"crypto/md5"
	b64 "encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"yunion.io/x/ocadm/pkg/apis/constants"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"

	// phases_init "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/init"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	staticpodutil "k8s.io/kubernetes/cmd/kubeadm/app/util/staticpod"
	"k8s.io/kubernetes/pkg/util/normalizer"
	ocadm_defaults "yunion.io/x/ocadm/pkg/apis/v1"
	ocadm_init "yunion.io/x/ocadm/pkg/phases/init"
	ocadm_join "yunion.io/x/ocadm/pkg/phases/join"
)

var (
	keepalivedLocalExample = normalizer.Examples(`
		# Generates the static Pod manifest file for Keepalived, functionally
		# equivalent to what is generated by kubeadm init.
		kubeadm init phase Keepalived local

		# Generates the static Pod manifest file for Keepalived using options
		# read from a configuration file.
		kubeadm init phase Keepalived local --config config.yaml
    `)
)

func NewKeepalivedPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:  "keepalived",
		Short: "Generate static Pod manifest file for local keepalived",
		Long:  cmdutil.MacroCommandLongDescription,
		Phases: []workflow.Phase{
			newKeepalivedSubPhase(),
		},
	}
	return phase
}

func newKeepalivedSubPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "local",
		Short:        "Generate the static Pod manifest file for a local, single-node local etcd instance",
		Example:      keepalivedLocalExample,
		Run:          runKeepalivedPhaseLocal(),
		InheritFlags: getKeepalivedPhaseFlags(),
	}
	return phase
}

func getKeepalivedPhaseFlags() []string {
	flags := []string{
		options.CertificatesDir,
		options.CfgPath,
		options.ImageRepository,
	}
	return flags
}

func runKeepalivedPhaseLocal() func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		vip := ""
		role := "MASTER"
		idata, ok := c.(ocadm_init.InitData)
		keepalivedVersionTag := ""
		if !ok {
			jdata, ok := c.(ocadm_join.JoinData)
			if !ok {
				return errors.New("Keepalived phase invoked with an invalid data struct")
			}
			if jdata.Cfg().ControlPlane == nil {
				return nil
			}
			role = "BACKUP"
			vip = jdata.GetHighAvailabilityVIP()
			keepalivedVersionTag = jdata.GetKeepalivedVersionTag()
		} else {
			vip = idata.GetHighAvailabilityVIP()
			keepalivedVersionTag = idata.GetKeepalivedVersionTag()
		}
		if len(vip) == 0 {
			fmt.Println("vip is empty. no need to install keepalived.")
			return nil
		}
		if len(keepalivedVersionTag) == 0 {
			fmt.Println("Keepalived version tag is empty! using default option: ", constants.DefaultKeepalivedVersionTag)
			keepalivedVersionTag = constants.DefaultKeepalivedVersionTag
		} else {
			fmt.Println("got Keepalived version tag from commandline: ", keepalivedVersionTag)
		}

		fmt.Printf("[PASS] Installing Keepalived:%s as %s", keepalivedVersionTag, role)
		dataPath := "/var/lib/keepalived"
		if err := os.MkdirAll(dataPath, 0700); err != nil {
			return errors.Wrapf(err, "failed to create Keepalived directory %q", dataPath)
		} else {
			fmt.Println("[PASS] keepalived path created.")
		}
		if err := CreateLocalKeepalivedStaticPodManifestFile(vip, keepalivedVersionTag, role); err != nil {
			return errors.Wrap(err, "error creating local keepalived static pod manifest file")
		}
		return nil
	}
}

func CreateLocalKeepalivedStaticPodManifestFile(ip, keepalivedVersionTag, role string) error {
	spec := GetKeepalivedPodSpec(ip, keepalivedVersionTag, role)
	if err := staticpodutil.WriteStaticPodToDisk("keepalived", "/etc/kubernetes/manifests", spec); err != nil {
		return err
	}
	return nil
}

func md5ize(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func GetKeepalivedPodSpec(vip, keepalivedVersionTag, role string) v1.Pod {
	privileged := true
	priority := "100"
	vipPswd := md5ize(b64.StdEncoding.EncodeToString([]byte(vip)))
	if len(vipPswd) > 8 {
		vipPswd = vipPswd[0:8]
	}
	// registry.cn-beijing.aliyuncs.com/yunionio/keepalived:v2.0.22
	containerName := "keepalived"
	imageUrl := fmt.Sprintf("%s/%s:%s", ocadm_defaults.DefaultImageRepository, containerName, keepalivedVersionTag)
	if role != "MASTER" {
		priority = "90"
	}

	vid, err := strconv.Atoi(strings.ReplaceAll(vip, ".", ""))
	if err == nil {
		fmt.Println(vid)
	}

	vid = vid % 255
	svid := fmt.Sprintf("%d", vid)

	return staticpodutil.ComponentPod(v1.Container{
		Name:            containerName,
		Command:         []string{"/container/tool/run"},
		Image:           imageUrl,
		ImagePullPolicy: v1.PullIfNotPresent,
		SecurityContext: &v1.SecurityContext{
			Privileged: &privileged,
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{
					v1.Capability("SYS_NICE"),
					v1.Capability("NET_ADMIN"),
					v1.Capability("NET_BROADCAST"),
					v1.Capability("NET_RAW"),
				},
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "KEEPALIVED_PRIORITY",
				Value: priority,
			},
			{
				Name:  "KEEPALIVED_VIRTUAL_IPS",
				Value: fmt.Sprintf("#PYTHON2BASH:['%s']", vip),
			},
			{
				Name:  "KEEPALIVED_STATE",
				Value: role,
			},
			{
				Name:  "KEEPALIVED_PASSWORD",
				Value: vipPswd,
			},
			{
				Name:  "KEEPALIVED_ROUTER_ID",
				Value: svid,
			},
		},
	}, nil)
}