package service

import "errors"

var (
	// ErrEmptyURL возвращается, когда передан пустой URL.
	ErrEmptyURL = errors.New("URL не может быть пустым")

	// ErrEmptyList возвращается, когда передан пустой список.
	ErrEmptyList = errors.New("список не может быть пустым")

	// ErrDBNotConfigured возвращается, когда база данных не настроена.
	ErrDBNotConfigured = errors.New("база данных не настроена")
)
