package storage

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("user already exists")
    ErrEventNotFound = errors.New("event not found")
    ErrEventExists   = errors.New("event already exists")
)