// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"os"
	"strings"
	"time"

	"fybrik.io/fybrik/manager/controllers/utils"
)

type Subscription struct {
	Entity       Subscriber
	Options      FileMonitorOptions
	LastModified time.Time
	NumFiles     int
}

// FileMonitor detects  changes since the last check and notifies subscribers
type FileMonitor struct {
	// list of subscribers and their states
	Subsciptions []Subscription
}

type FileMonitorOptions struct {
	// Directory that should be monitored for change
	Path string
	// File extension
	Extension string
}

func (m *FileMonitor) Subscribe(subscriber Subscriber) error {
	s := Subscription{
		Entity:  subscriber,
		Options: subscriber.GetOptions(),
	}
	numFiles, modified, err := m.visit(&s)
	if err != nil {
		return err
	}
	s.LastModified = modified
	s.NumFiles = numFiles
	m.Subsciptions = append(m.Subsciptions, s)
	return nil
}

func (m *FileMonitor) Run() {
	go func() {
		for {
			m.Monitor()
			time.Sleep(utils.GetMonitorInterval())
		}
	}()
}

func (m *FileMonitor) Monitor() {
	for i := range m.Subsciptions {
		s := &m.Subsciptions[i]
		numFiles, lastTimestamp, err := m.visit(s)
		if err != nil {
			s.Entity.OnError(err)
			return
		}
		changeDetected := false
		if numFiles != s.NumFiles {
			s.NumFiles = numFiles
			changeDetected = true
		}
		if lastTimestamp.After(s.LastModified) {
			s.LastModified = lastTimestamp
			changeDetected = true
		}
		if changeDetected {
			s.Entity.OnNotify()
		}
	}
}

func (m *FileMonitor) visit(s *Subscription) (int, time.Time, error) {
	modified := s.LastModified
	numFiles := 0
	entries, err := os.ReadDir(s.Options.Path)
	if err != nil {
		return numFiles, modified, err
	}
	for _, entry := range entries {
		info, _ := entry.Info()
		if info.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), s.Options.Extension) {
			continue
		}
		numFiles++
		modTime := info.ModTime()
		if info.Mode()&os.ModeSymlink != 0 {
			absName, err := os.Readlink(s.Options.Path + "/" + info.Name())
			if err != nil {
				return numFiles, modified, err
			}
			absEntry, err := os.Lstat(s.Options.Path + "/" + absName)
			if err != nil {
				return numFiles, modified, err
			}
			modTime = absEntry.ModTime()
		}

		if modTime.After(modified) {
			modified = modTime
		}
	}
	return numFiles, modified, nil
}
