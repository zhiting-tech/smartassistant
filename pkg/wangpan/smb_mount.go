package wangpan

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/utils/exec"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"golang.org/x/crypto/md4"
	"golang.org/x/sync/errgroup"
)

type SmbMountStr struct {
	echo               exec.ECmd
	sed                exec.ECmd
	CurrentAccountName string
	AccountName        string
	Password           string
	defMountPath       string
}

const (
	PASSWD    string = "%s:x:1001:1001::/home/%s:/bin/ash"
	SHADOW    string = "%s:!:19053::::::"
	SMBPassWd string = "%s:1001:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX:%s:[U          ]:LCT-621EE007:"
)

func NewSmbMountStr(currentAccountName, accountName, passwd string) *SmbMountStr {
	return &SmbMountStr{
		echo:               exec.GetCmd("echo"),
		sed:                exec.GetCmd("sed"),
		CurrentAccountName: currentAccountName,
		AccountName:        accountName,
		Password:           passwd,
	}
}

func (smb *SmbMountStr) SetMountPath(path string) error {
	if path == "" {
		path = filepath.Join(config.GetConf().SmartAssistant.RuntimePath, "run", "wangpan")
	}
	if _, err := os.Stat(path); err != nil {
		if err = os.Mkdir(path, 0777); err != nil {
			logger.Error("SetMountPath err:", err)
			return err
		}
	}
	smb.defMountPath = path
	return nil
}

func (smb *SmbMountStr) mountPath(fileName string) string {
	return filepath.Join(smb.defMountPath, fileName)
}

func (smb *SmbMountStr) Exec() error {
	if smb.AccountName != "" && smb.Password != "" {
		return smb.configFileProcess(smb.addAccountConf)
	} else if (smb.AccountName != "" && smb.Password == "") || (smb.AccountName == "" && smb.Password != "") {
		return smb.configFileProcess(smb.updateAccountConf)
	}
	return smb.configFileProcess(smb.delAccountConf)

}

func (smb *SmbMountStr) configFileProcess(f func() ([][]string, int, error)) error {
	group := new(errgroup.Group)
	fields, ty, err := f()
	if err != nil {
		return err
	}
	for _, v := range fields {
		arg := v
		if arg == nil {
			continue
		}
		group.Go(func() error {
			switch ty {
			case 1:
				err = smb.sed.Exec(arg...).ExecRedirect()
			case 2:
				err = smb.echo.Exec(arg...).ExecOutPut()
			}
			return err
		})
	}
	if err = group.Wait(); err != nil {
		return err
	}
	return nil
}

func (smb *SmbMountStr) initAccountConf() ([][]string, int, error) {
	var (
		fields = make([][]string, 3, 3)
	)
	passwd := fmt.Sprintf(PASSWD, "nobody", "nobody")
	shadow := fmt.Sprintf(SHADOW, "nobody")
	smbPassWd := fmt.Sprintf(SMBPassWd, "nobody", GetUcs2Md4("nobody"))

	fields[0] = []string{
		"-c",
		fmt.Sprintf("echo %s > %s", passwd, smb.mountPath("passwd")),
	}
	fields[1] = []string{
		"-c",
		fmt.Sprintf("echo %s > %s", shadow, smb.mountPath("shadow")),
	}
	fields[2] = []string{
		"-c",
		fmt.Sprintf("echo %s > %s", smbPassWd, smb.mountPath("smbpasswd")),
	}
	return fields, 2, nil
}

func (smb *SmbMountStr) addAccountConf() ([][]string, int, error) {
	var (
		fields = make([][]string, 3, 3)
	)

	if _, err := os.Stat(smb.mountPath("passwd")); err != nil {
		if err = smb.configFileProcess(smb.initAccountConf); err != nil {
			return nil, 0, err
		}
	}

	passwd := fmt.Sprintf(PASSWD, smb.AccountName, smb.AccountName)
	shadow := fmt.Sprintf(SHADOW, smb.AccountName)
	smbPassWd := fmt.Sprintf(SMBPassWd, smb.AccountName, GetUcs2Md4(smb.Password))
	// file locate
	fields[0] = []string{
		fmt.Sprintf("1i%s", passwd), smb.mountPath("passwd"),
	}
	fields[1] = []string{
		fmt.Sprintf("1i%s", shadow), smb.mountPath("shadow"),
	}

	fields[2] = []string{
		fmt.Sprintf("1i%s", smbPassWd), smb.mountPath("smbpasswd"),
	}
	return fields, 1, nil
}

func (smb *SmbMountStr) updateAccountConf() ([][]string, int, error) {
	var (
		fields = make([][]string, 3, 3)
	)
	if smb.AccountName == "" {
		fields[0] = []string{
			"-e", fmt.Sprintf("1i%s", fmt.Sprintf(SMBPassWd, smb.CurrentAccountName, GetUcs2Md4(smb.Password))),
			"-e", fmt.Sprintf("/%s:/d", smb.CurrentAccountName), smb.mountPath("smbpasswd"),
		}

	} else {
		fields[0] = []string{fmt.Sprintf("s/%s:x/%s:x/g", smb.CurrentAccountName, smb.AccountName), smb.mountPath("passwd")}
		fields[1] = []string{fmt.Sprintf("s/%s:!/%s:!/g", smb.CurrentAccountName, smb.AccountName), smb.mountPath("shadow")}
		fields[2] = []string{fmt.Sprintf("s/%s:/%s:/g", smb.CurrentAccountName, smb.AccountName), smb.mountPath("smbpasswd")}
	}
	return fields, 1, nil
}

func (smb *SmbMountStr) delAccountConf() ([][]string, int, error) {
	var (
		fields = make([][]string, 3, 3)
	)
	fields[0] = []string{fmt.Sprintf("/%s:x/d", smb.CurrentAccountName), smb.mountPath("passwd")}
	fields[1] = []string{fmt.Sprintf("/%s:!/d", smb.CurrentAccountName), smb.mountPath("shadow")}
	fields[2] = []string{fmt.Sprintf("/%s:/d", smb.CurrentAccountName), smb.mountPath("smbpasswd")}
	return fields, 1, nil
}

func StringToUcs2(str string) []byte {
	if str == "" {
		return nil
	}
	u := utf16.Encode([]rune(str))
	dst := make([]byte, len(u)*2)
	wi := 0
	for _, r := range u {
		binary.LittleEndian.PutUint16(dst[wi:], uint16(r))
		wi += 2
	}
	return dst
}
func GetMd4(str string) string {
	md4New := md4.New()
	md4New.Write([]byte(str))
	md4String := hex.EncodeToString(md4New.Sum(nil))
	return strings.ToUpper(md4String)
}

func GetUcs2Md4(str string) string {
	return GetMd4(string(StringToUcs2(str)))
}
