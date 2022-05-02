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
	numFiles, modified, err := m.visit(&s, s.Options.Path)
	if err != nil {
		return err
	}
	s.LastModified = modified
	s.NumFiles = numFiles
	m.Subsciptions = append(m.Subsciptions, s)
	return nil
}

func (m *FileMonitor) Run() {
	ticker := time.NewTicker(utils.GetMonitorInterval())
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			m.Monitor()
		}
	}()
}

func (m *FileMonitor) Monitor() {
	for i := range m.Subsciptions {
		s := &m.Subsciptions[i]
		numFiles, lastTimestamp, err := m.visit(s, s.Options.Path)
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

func (m *FileMonitor) visit(s *Subscription, path string) (int, time.Time, error) {
	modified := s.LastModified
	numFiles := 0
	entries, err := os.ReadDir(s.Options.Path)
	if err != nil {
		return numFiles, modified, err
	}
	for _, entry := range entries {
		info, _ := entry.Info()
		if !info.IsDir() {
			if !strings.HasSuffix(entry.Name(), s.Options.Extension) {
				continue
			}
			numFiles++
			if info.ModTime().After(modified) {
				modified = info.ModTime()
			}
		} else {
			dirNumFiles, dirModified, err := m.visit(s, path+"/"+info.Name())
			if err != nil {
				return numFiles, modified, err
			}
			numFiles += dirNumFiles
			if dirModified.After(modified) {
				modified = dirModified
			}
		}
	}
	return numFiles, modified, nil
}
