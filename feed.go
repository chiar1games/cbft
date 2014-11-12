//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package main

type Feed interface {
	Name() string
	Start() error
	Close() error
	Dests() map[string]Dest // Key is partition identifier.
}

// Default values for feed parameters.
const FEED_SLEEP_MAX_MS = 10000
const FEED_SLEEP_INIT_MS = 100
const FEED_BACKOFF_FACTOR = 1.5

type FeedStartFunc func(mgr *Manager, feedName, indexName, indexUUID string,
	sourceType, sourceName, sourceUUID, sourceParams string,
	dests map[string]Dest) error

var feedTypes = make(map[string]FeedStartFunc)

func RegisterFeedType(sourceType string, fn FeedStartFunc) {
	feedTypes[sourceType] = fn
}
