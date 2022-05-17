// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

type Subscription struct {
	Entity       Subscriber
	Options      FileMonitorOptions
	LastModified time.Time
	NumFiles     int
}

// FileMonitor detects  changes since the last check and notifies subscribers
type FileMonitor struct {
	// log
	Log zerolog.Logger
	// list of subscribers and their states
	Subsciptions []Subscription
}

type FileMonitorOptions struct {
	// Directory that should be monitored for change
	Path string
	// File extension
	Extension string
}

// Subscribe to get notifications on file changes
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

// Run activates a go routine that calls Monitor() upon changes in /tmp/adminconfig directory
func (m *FileMonitor) Run(watcher *fsnotify.Watcher) {
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				// not watching chmod event
				if event.Op == fsnotify.Chmod {
					continue
				}
				m.Log.Info().Msg("Event: " + event.String())
				m.Monitor()
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				m.Log.Err(err).Msg("error watching file changes")
			}
		}
	}()
}

// Check if changes in files require any action
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
			// a file has been added or removed
			s.NumFiles = numFiles
			changeDetected = true
		}
		if lastTimestamp.After(s.LastModified) {
			// a file has been modified
			s.LastModified = lastTimestamp
			changeDetected = true
		}
		if changeDetected {
			m.Log.Info().Msg("File change detected, notifying...")
			s.Entity.OnNotify()
		}
	}
}

// Scan the montored folder to update the following:
// - number of files with the relevant extension
// - the latest timestamp of these files
func (m *FileMonitor) visit(s *Subscription) (int, time.Time, error) {
	modified := s.LastModified
	numFiles := 0
	entries, err := os.ReadDir(s.Options.Path)
	if err != nil {
		return 0, s.LastModified, err
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return 0, s.LastModified, err
		}
		if info.IsDir() {
			// only policy/infrastructure files are monitored
			continue
		}
		if !strings.HasSuffix(entry.Name(), s.Options.Extension) {
			// this file should not be monitored for the given subscriber
			continue
		}
		numFiles++
		modTime := info.ModTime()
		if info.Mode()&os.ModeSymlink != 0 {
			// symbolic link
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
