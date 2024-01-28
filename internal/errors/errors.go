package errors

import "errors"

var ErrUnknownDeviceType = errors.New("unknown device type")
var ErrTeltonikaUnauthorizedDevice = errors.New("unauthorized teltonika device")
var ErrTeltonikaInvalidDataPacket = errors.New("invalid teltonika data packet")
var ErrTeltonikaBadCrc = errors.New("crc check failed for teltonika device")

var ErrNotWanwayDevice = errors.New("not a wanway device")
var ErrWanwayInvalidPacket = errors.New("invalid wanway data packet")
var ErrWanwayInvalidLoginInfo = errors.New("invalid wanway login info")
var ErrWanwayBadCrc = errors.New("crc check failed for wanway device")
var ErrWanwayUnauthorizedDevice = errors.New("unauthorized wanway device")
