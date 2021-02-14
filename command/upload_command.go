package command

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/linklux/luxbox-client/component"

	"github.com/cheggaaa/pb/v3"
)

const CHUNK_SIZE = 1024

type UploadCommand struct {
	command

	component.ServerConnector
}

func (cmd UploadCommand) New() ICommand {
	return UploadCommand{
		command{
			"upload",
			"Copy the resource to the Luxbox server",
			map[string]*commandFlag{
				"name":      &commandFlag{"name", "n", "string", "Create the resource on the server with the specified name, omit to use local resource name", ""},
				"overwrite": &commandFlag{"overwrite", "o", "bool", "Overwrite resource if it already exists", false},
			},
		},
		component.ServerConnector{},
	}
}

func (cmd UploadCommand) Execute(args []string) error {
	if len(args) < 1 {
		return errors.New("missing name or path to the local resource")
	}

	// TODO Support multifile upload
	workingDir, _ := os.Getwd()
	absPath := path.Join(workingDir, args[0])

	// Open file and collect required file meta
	file, err := os.Open(absPath)
	if err != nil {
		return errors.New("failed to open file: " + err.Error())
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return errors.New("failed to retrieve file info: " + err.Error())
	}

	// Determine resourcename, use name flag if provided, local filename otherwise
	resourceName := filepath.Base(args[0])
	if nameFlag := cmd.flags["name"].Value.(string); nameFlag != "" {
		resourceName = nameFlag
	}

	request := component.Request{
		Action: "upload",
		Meta: map[string]interface{}{
			"resourceSize": fi.Size(),
			"resourceName": resourceName,
			"overwrite":    cmd.flags["overwrite"].Value.(bool),
		},
	}

	cmd.Connect()
	cmd.UserAuthEnabled(true)
	defer cmd.Disconnect()

	err = cmd.SendRequest(request)
	if err != nil {
		return err
	}

	// TODO Catch error responses
	// Wait until the server is ready to receive the data stream
	err = cmd.WaitForMessage("ready")
	if err != nil {
		return err
	}

	// Server is ready, start streaming data
	putFile(file, cmd.GetConnection(), fi.Size(), resourceName)

	response, err := cmd.GetResponse()
	if err != nil {
		return err
	} else if response.Code != 3 {
		return errors.New(fmt.Sprintf("unexpected response code %d, response %v", response.Code, response.Data))
	}

	return nil
}

func putFile(file *os.File, conn net.Conn, size int64, name string) error {
	r := bufio.NewReader(file)
	read, written, chunks := int64(0), int64(0), int(0)

	buf := make([]byte, 0, 1024)

	// progress bar
	template := `{{string . "filename"}}  {{speed . "%8s/s" "       -/s"}}  {{counters . "%8s / %8s" "%s/-       "}} {{etime . "%4s"}} {{bar . "[" "#" "#" " " "]"}} {{percent . "%6.02f%%" "?"}}`
	bar := pb.ProgressBarTemplate(template).Start64(size)
	barReader := bar.NewProxyReader(r)

	bar.Set("filename", fmt.Sprintf("%-30s", name))

	for read < size {
		n, err := barReader.Read(buf[:cap(buf)])
		if n < 1 {
			continue
		}

		chunks++
		read += int64(n)

		if err != nil {
			fmt.Printf("error during buffer read: " + err.Error())
			break
		}

		n, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Printf("error during file write: " + err.Error())
			break
		}

		written += int64(n)

		// When reaching the last chunk, resize the buffer to match
		if size-read < CHUNK_SIZE && size-read > 0 {
			buf = make([]byte, 0, size-read)
		}

		// Used for local development, testing is annoying when something
		// related to networking is instant.
		time.Sleep(500000)
	}

	bar.Finish()

	if read != written {
		return errors.New(fmt.Sprintf("byte count read from file (%d) does not match bytes written to server (%d)", read, written))
	}

	return nil
}
