package ai

import "errors"

var (
	ErrProviderUnavailable = errors.New("penyedia model tidak tersedia")
	ErrProviderRejected    = errors.New("penyedia model menolak permintaan")
	ErrTimeout             = errors.New("waktu permintaan penyedia model habis")
	ErrInvalidResponse     = errors.New("respons penyedia model tidak sah")
	ErrRequestTooLarge     = errors.New("permintaan penyedia model terlalu besar")
	ErrResponseTooLarge    = errors.New("respons penyedia model terlalu besar")
)
