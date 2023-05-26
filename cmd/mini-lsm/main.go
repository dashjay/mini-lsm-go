package main

import (
	"flag"
	"io"
	"strings"
	"unsafe"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/redcon"

	"github.com/dashjay/mini-lsm-go/pkg/lsm"
)

func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

func handleCmd(c redcon.DetachedConn, cmd redcon.Command, kv *lsm.Storage) {
	defer c.Flush()
	switch strings.ToLower(string(cmd.Args[0])) {
	case "get":
		if len(cmd.Args) > 2 {
			c.WriteError("ERR wrong number of arguments")
			return
		}
		out := kv.Get(cmd.Args[1])
		if len(out) != 0 {
			c.WriteString(b2s(out))
		} else {
			c.WriteError("ERR Notfound")
		}
		err := c.Flush()
		if err != nil {
			_ = c.Close()
		}
		return
	case "set":
		if len(cmd.Args) > 3 {
			c.WriteError("ERR wrong number of arguments")
			return
		}
		kv.Put(cmd.Args[1], cmd.Args[2])
		c.WriteString("OK")
	default:
		logrus.Errorf("unknown command: %s", string(cmd.Args[0]))
		c.WriteError("ERR unknown command " + string(cmd.Args[0]))
	}
}

func main() {
	var dir = flag.String("dir", "", "dir for use")
	var bind = flag.String("bind", ":8080", "listen port")

	flag.Parse()
	if *dir == "" {
		logrus.Fatalf("--dir should be set for workdir")
	}
	lsmKV := lsm.NewStorage(*dir)
	server := redcon.NewServer(*bind,
		func(conn redcon.Conn, cmd redcon.Command) {
			dc := conn.Detach()
			defer dc.Close()
			for {
				handleCmd(dc, cmd, lsmKV)
				var err error
				cmd, err = dc.ReadCommand()
				if err != nil {
					if err != io.EOF {
						dc.WriteError("ERR " + err.Error())
					}
					dc.Flush()
					return
				}
			}
		},
		func(conn redcon.Conn) bool {
			return true
		},
		func(conn redcon.Conn, err error) {
			logrus.WithError(err).WithField("remote_addr", conn.RemoteAddr()).Debugln("remote connect closed")
		},
	)
	panic(server.ListenAndServe())
}
