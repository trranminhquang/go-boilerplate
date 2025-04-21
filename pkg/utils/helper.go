package utils

import "github.com/sirupsen/logrus"

func Must[T any](val T, err error) T {
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get value")
	}
	return val
}
