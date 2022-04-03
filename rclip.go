// rclip
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/washtubs/rclip/common"
	"github.com/washtubs/rclip/receiver"
	"github.com/washtubs/rclip/reflector"
	"github.com/washtubs/rclip/sender"
)

func main() {
	var ver = 1.2
	version_str := fmt.Sprintf("remote clipboard, version %v", ver)
	fmt.Println(version_str)
	role := flag.String("role", "rcv", "specify the role, rcv/refl/sender") //refl will accept text from sender, and forward it to rcv
	send_port := flag.Uint("send_port", 8890, "specify the port for sender on reflector")
	rcv_port := flag.Uint("rcv_port", 8891, "specify the port for receiver on reflector")
	refl_ip := flag.String("refl_ip", "", "reflector listening ip address")
	loose_san_check := flag.Bool("loose", false, "use hard coded address for reflector certificate SAN check")
	eventName := flag.String("event", "", "Event name (receiver only)")
	flag.Parse()
	conf_dir := common.GetConfDir()
	switch *role {
	case "refl":
		refl, err := reflector.NewReflector(filepath.Join(conf_dir, "ca_cert"), filepath.Join(conf_dir, "refl_cert"), filepath.Join(conf_dir, "refl_key"), *refl_ip, int(*send_port), int(*rcv_port))
		if err != nil {
			common.ErrLog.Println(err)
			return
		}
		refl.Start()
	case "rcv":
		rcv, err := receiver.NewRclipRcv(filepath.Join(conf_dir, "ca_cert"), filepath.Join(conf_dir, "rcv_cert"), filepath.Join(conf_dir, "rcv_key"), *refl_ip, int(*rcv_port), !*loose_san_check)
		if err != nil {
			common.ErrLog.Println(err)
			return
		}

		stat, _ := os.Stdin.Stat()
		stdin := []byte{}
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdin, err = io.ReadAll(os.Stdin)
			if err != nil {
				common.ErrLog.Println(err)
				return
			}
		}
		event := common.Event{
			Name:  *eventName,
			Args:  flag.Args(),
			Stdin: stdin,
		}
		rcv.ReadandSend(event)

	case "sender":
		snd, err := sender.NewRclipSender(filepath.Join(conf_dir, "ca_cert"), filepath.Join(conf_dir, "sender_cert"), filepath.Join(conf_dir, "sender_key"), *refl_ip, int(*send_port), !*loose_san_check)
		if err != nil {
			common.ErrLog.Println(err)
			return
		}
		snd.Receive()

	default:
		common.ErrLog.Println("wrong role type, it need to be one of rcv/refl/sender")
		return
	}
}
