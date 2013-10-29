// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"testing"
	native_time "time"
)

func TestComparators(t *testing.T) {
	t1a := TimestampFromUnix(0)
	t1b := TimestampFromUnix(0)
	t2 := TimestampFromUnix(2*int64(Second) - 1)

	if !t1a.Equal(t1b) {
		t.Fatalf("Expected %s to be equal to %s", t1a, t1b)
	}
	if t1a.Equal(t2) {
		t.Fatalf("Expected %s to not be equal to %s", t1a, t2)
	}

	if !t1a.Before(t2) {
		t.Fatalf("Expected %s to be before %s", t1a, t2)
	}
	if t1a.Before(t1b) {
		t.Fatalf("Expected %s to not be before %s", t1a, t1b)
	}

	if !t2.After(t1a) {
		t.Fatalf("Expected %s to be after %s", t2, t1a)
	}
	if t1b.After(t1a) {
		t.Fatalf("Expected %s to not be after %s", t1b, t1a)
	}
}

func TestTimestampConversions(t *testing.T) {
	unix := int64(1136239445)
	t1 := native_time.Unix(unix, 0)
	t2 := native_time.Unix(unix, int64(Second)-1)

	ts := TimestampFromUnix(unix)
	if !ts.Time().Equal(t1) {
		t.Fatalf("Expected %s, got %s", t1, ts.Time())
	}

	// Test available precision.
	ts = TimestampFromTime(t2)
	if !ts.Time().Equal(t1) {
		t.Fatalf("Expected %s, got %s", t1, ts.Time())
	}

	if ts.Unix() != unix {
		t.Fatalf("Expected %d, got %d", unix, ts.Unix())
	}
}

func TestDuration(t *testing.T) {
	scenarios := []struct {
		timeDuration native_time.Duration
		duration     Duration
	}{
		{
			timeDuration: native_time.Second,
			duration:     Second,
		},
		{
			timeDuration: native_time.Minute,
			duration:     Minute,
		},
		{
			timeDuration: native_time.Hour,
			duration:     Hour,
		},
	}

	for i, s := range scenarios {
		if NewDuration(s.timeDuration) != s.duration {
			t.Fatalf("%d. Expected %d, got %d", i, s.duration, NewDuration(s.timeDuration))
		}
		if s.duration.TimeDuration() != s.timeDuration {
			t.Fatalf("%d. Expected %d, got %d", i, s.duration.TimeDuration(), s.timeDuration)
		}

		goTime := native_time.Unix(1136239445, 0)
		ts := TimestampFromTime(goTime)
		if !goTime.Add(s.timeDuration).Equal(ts.Add(s.duration).Time()) {
			t.Fatalf("%d. Expected %s to be equal to %s", goTime.Add(s.timeDuration), ts.Add(s.duration))
		}

		earlier := ts.Add(-s.duration)
		delta := ts.Sub(earlier)
		if delta.TimeDuration() != s.timeDuration {
			t.Fatalf("%d. Expected %s to be equal to %s", delta.TimeDuration, s.timeDuration)
		}
	}
}
