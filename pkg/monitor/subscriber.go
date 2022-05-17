// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package monitor

type Subscriber interface {
	GetOptions() FileMonitorOptions
	OnError(err error)
	OnNotify()
}
