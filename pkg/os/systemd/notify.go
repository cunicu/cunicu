// SPDX-FileCopyrightText: 2014 Docker, Inc.
// SPDX-FileCopyrightText: 2015-2018 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package systemd

const (
	// NotifyReady tells the service manager that service startup is finished
	// or the service finished loading its configuration.
	NotifyReady = "READY=1"

	// NotifyStopping tells the service manager that the service is beginning
	// its shutdown.
	NotifyStopping = "STOPPING=1"

	// NotifyReloading tells the service manager that this service is
	// reloading its configuration. Note that you must call SdNotifyReady when
	// it completed reloading.
	NotifyReloading = "RELOADING=1"

	// NotifyWatchdog tells the service manager to update the watchdog
	// timestamp for the service.
	NotifyWatchdog = "WATCHDOG=1"
)
