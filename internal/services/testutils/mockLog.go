package testutils

import "github.com/Tbits007/auth/internal/lib/logger/slogdiscard"

var (
	Log = slogdiscard.NewDiscardLogger()
)