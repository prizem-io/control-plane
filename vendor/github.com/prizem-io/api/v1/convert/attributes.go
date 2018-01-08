// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"time"

	"github.com/prizem-io/api/v1/proto"
)

// EncodeAttributes encodes a native Go map to a `proto.Attributes`.
func EncodeAttributes(md map[string]interface{}) proto.Attributes {
	var attrs proto.Attributes

	if md == nil {
		return attrs
	}

	var index int32
	words := make([]string, 0, len(md)+10)

	for k, v := range md {
		words, index = wordIndex(words, k)
		switch t := v.(type) {
		case string:
			if attrs.Strings == nil {
				attrs.Strings = make(map[int32]int32)
			}
			var ti int32
			words, ti = wordIndex(words, t)
			attrs.Strings[index] = ti
		case []string:
			if attrs.StringList == nil {
				attrs.StringList = make(map[int32]proto.StringList)
			}
			items := make([]int32, len(t))
			for i, v := range t {
				var ti int32
				words, ti = wordIndex(words, v)
				items[i] = ti
			}
			attrs.StringList[index] = proto.StringList{
				Items: items,
			}
		case int:
			ensureInt64s(&attrs)
			attrs.Int64S[index] = int64(t)
		case int8:
			ensureInt64s(&attrs)
			attrs.Int64S[index] = int64(t)
		case int16:
			ensureInt64s(&attrs)
			attrs.Int64S[index] = int64(t)
		case int32:
			ensureInt64s(&attrs)
			attrs.Int64S[index] = int64(t)
		case int64:
			ensureInt64s(&attrs)
			attrs.Int64S[index] = t
		case []int:
			ensureInt64List(&attrs)
			items := make([]int64, len(t))
			for i, v := range t {
				items[i] = int64(v)
			}
			attrs.Int64List[index] = proto.Int64List{
				Items: items,
			}
		case []int8:
			ensureInt64List(&attrs)
			items := make([]int64, len(t))
			for i, v := range t {
				items[i] = int64(v)
			}
			attrs.Int64List[index] = proto.Int64List{
				Items: items,
			}
		case []int16:
			ensureInt64List(&attrs)
			items := make([]int64, len(t))
			for i, v := range t {
				items[i] = int64(v)
			}
			attrs.Int64List[index] = proto.Int64List{
				Items: items,
			}
		case []int32:
			ensureInt64List(&attrs)
			items := make([]int64, len(t))
			for i, v := range t {
				items[i] = int64(v)
			}
			attrs.Int64List[index] = proto.Int64List{
				Items: items,
			}
		case []int64:
			ensureInt64List(&attrs)
			attrs.Int64List[index] = proto.Int64List{
				Items: t,
			}
		case float32:
			ensureDoubles(&attrs)
			attrs.Doubles[index] = float64(t)
		case float64:
			ensureDoubles(&attrs)
			attrs.Doubles[index] = t
		case []float32:
			ensureDoubleList(&attrs)
			items := make([]float64, len(t))
			for i, v := range t {
				items[i] = float64(v)
			}
			attrs.DoubleList[index] = proto.DoubleList{
				Items: items,
			}
		case []float64:
			ensureDoubleList(&attrs)
			attrs.DoubleList[index] = proto.DoubleList{
				Items: t,
			}
		case bool:
			if attrs.Bools == nil {
				attrs.Bools = make(map[int32]bool)
			}
			attrs.Bools[index] = t
		case []bool:
			if attrs.BoolList == nil {
				attrs.BoolList = make(map[int32]proto.BoolList)
			}
			attrs.BoolList[index] = proto.BoolList{
				Items: t,
			}
		case time.Time:
			if attrs.Timestamps == nil {
				attrs.Timestamps = make(map[int32]time.Time)
			}
			attrs.Timestamps[index] = t
		case []time.Time:
			if attrs.TimestampList == nil {
				attrs.TimestampList = make(map[int32]proto.TimestampList)
			}
			attrs.TimestampList[index] = proto.TimestampList{
				Items: t,
			}
		case time.Duration:
			if attrs.Durations == nil {
				attrs.Durations = make(map[int32]time.Duration)
			}
			attrs.Durations[index] = t
		case []time.Duration:
			if attrs.DurationList == nil {
				attrs.DurationList = make(map[int32]proto.DurationList)
			}
			attrs.DurationList[index] = proto.DurationList{
				Items: t,
			}
		case []byte:
			if attrs.Bytes == nil {
				attrs.Bytes = make(map[int32][]byte)
			}
			attrs.Bytes[index] = t
		case map[string]interface{}:
			if attrs.Attributes == nil {
				attrs.Attributes = make(map[int32]proto.Attributes)
			}
			attrs.Attributes[index] = EncodeAttributes(t)
		}
	}

	attrs.Words = words

	return attrs
}

// DecodeAttributes decodes a `proto.Attributes` to a native Go map.
func DecodeAttributes(attrs proto.Attributes) map[string]interface{} {
	fields := make(map[string]interface{})
	for k, v := range attrs.Strings {
		fields[attrs.Words[k]] = attrs.Words[v]
	}
	for k, v := range attrs.StringList {
		vals := make([]string, len(v.Items))
		for i, val := range v.Items {
			vals[i] = attrs.Words[val]
		}
		fields[attrs.Words[k]] = vals
	}
	for k, v := range attrs.Int64S {
		fields[attrs.Words[k]] = v
	}
	for k, v := range attrs.Int64List {
		fields[attrs.Words[k]] = v.Items
	}
	for k, v := range attrs.Doubles {
		fields[attrs.Words[k]] = v
	}
	for k, v := range attrs.DoubleList {
		fields[attrs.Words[k]] = v.Items
	}
	for k, v := range attrs.Bools {
		fields[attrs.Words[k]] = v
	}
	for k, v := range attrs.BoolList {
		fields[attrs.Words[k]] = v.Items
	}
	for k, v := range attrs.Timestamps {
		fields[attrs.Words[k]] = v
	}
	for k, v := range attrs.TimestampList {
		fields[attrs.Words[k]] = v.Items
	}
	for k, v := range attrs.Durations {
		fields[attrs.Words[k]] = v
	}
	for k, v := range attrs.DurationList {
		fields[attrs.Words[k]] = v.Items
	}
	for k, v := range attrs.Bytes {
		fields[attrs.Words[k]] = v
	}
	for k, v := range attrs.Attributes {
		fields[attrs.Words[k]] = DecodeAttributes(v)
	}

	return fields
}

func wordIndex(words []string, word string) ([]string, int32) {
	for i, v := range words {
		if v == word {
			return words, int32(i)
		}
	}

	last := len(words)
	return append(words, word), int32(last)
}

func ensureInt64s(attrs *proto.Attributes) {
	if attrs.Int64S == nil {
		attrs.Int64S = make(map[int32]int64)
	}
}

func ensureInt64List(attrs *proto.Attributes) {
	if attrs.Int64List == nil {
		attrs.Int64List = make(map[int32]proto.Int64List)
	}
}

func ensureDoubles(attrs *proto.Attributes) {
	if attrs.Doubles == nil {
		attrs.Doubles = make(map[int32]float64)
	}
}

func ensureDoubleList(attrs *proto.Attributes) {
	if attrs.DoubleList == nil {
		attrs.DoubleList = make(map[int32]proto.DoubleList)
	}
}
