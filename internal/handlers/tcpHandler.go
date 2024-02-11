package handlers

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"path"

	devices "github.com/404minds/avl-receiver/internal/devices"
	errs "github.com/404minds/avl-receiver/internal/errors"
	configuredLogger "github.com/404minds/avl-receiver/internal/logger"
	"github.com/404minds/avl-receiver/internal/store"
)

var logger = configuredLogger.Logger

const BUFFER_SIZE = 256 // bytes

type tcpHandler struct {
	connToProtocolMap     map[string]devices.DeviceProtocol // make this an LRU cache to evict stale connections
	registeredDeviceTypes []devices.AVLDeviceType
	connToStoreMap        map[string]store.Store
	dataDir               string
}

func (t *tcpHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	deviceProtocol, ack, err := t.attemptDeviceLogin(reader)
	if err != nil {
		logger.Sugar().Errorf("failed to identify device from %s : %s", conn.RemoteAddr().String(), err)
		return
	}

	t.connToProtocolMap[conn.RemoteAddr().String()] = deviceProtocol
	dataStore := makeJsonStore(t.dataDir, deviceProtocol.GetDeviceIdentifier())
	go dataStore.Process()
	defer func() { dataStore.GetCloseChan() <- true }()

	t.connToStoreMap[conn.RemoteAddr().String()] = dataStore
	conn.Write(ack)

	writer := bufio.NewWriter(conn)
	err = deviceProtocol.ConsumeStream(reader, writer, dataStore.GetProcessChan())
	if err != nil && err != io.EOF {
		logger.Sugar().Errorf("Error reading from connection %s", conn.RemoteAddr().String())
		logger.Error(err.Error())
		return
	} else if err == io.EOF {
		logger.Sugar().Infof("Connection %s closed", conn.RemoteAddr().String())
		return
	}
}

func makeJsonStore(datadir string, deviceIdentifier string) store.Store {
	file, err := os.OpenFile(path.Join(datadir, deviceIdentifier+".json"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("failed to open file to store data")
		logger.Panic(err.Error())
	}
	logger.Sugar().Infof("[deviceId: %s] Created json file store at %s", deviceIdentifier, file.Name())

	return &store.JsonLinesStore{
		File:        file,
		ProcessChan: make(chan interface{}, 200),
		CloseChan:   make(chan bool, 200),
	}
}

func (t *tcpHandler) attemptDeviceLogin(reader *bufio.Reader) (devices.DeviceProtocol, []byte, error) {
	for _, deviceType := range t.registeredDeviceTypes {
		protocol := deviceType.GetProtocol()
		ack, bytesToSkip, err := protocol.Login(reader)

		if err != nil {
			if errors.Is(err, errs.ErrUnknownDeviceType) {
				continue // try another device
			} else {
				return nil, nil, err
			}
		} else {
			logger.Sugar().Infof("Device identified to be of type %s with identifier %s, bytes to skip %d", deviceType.String(), protocol.GetDeviceIdentifier(), bytesToSkip)
			if _, err := reader.Discard(bytesToSkip); err != nil {
				return nil, nil, err
			}
			return protocol, ack, nil
		}
	}

	return nil, nil, errs.ErrUnknownDeviceType
}
