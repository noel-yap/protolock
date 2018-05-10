package protolock

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const simpleProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
	rpc Next(stream NextRequest) returns (Channel);
	rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noUsingReservedFieldsProto = `syntax = "proto3";
package test;

message Channel {
  reserved 4, 8 to 11;
  reserved "foo", "bar";  
  int64 id = 1;
  string name = 2;
  string description = 3;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
	rpc Next(stream NextRequest) returns (Channel);
	rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const usingReservedFieldsProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noRemoveReservedFieldsProto = `syntax = "proto3";
package test;

message Channel {
  reserved 44, 101, 103 to 110;
  reserved "no_more", "goodbye";
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const removeReservedFieldsProto = `syntax = "proto3";
package test;

message Channel {
  reserved 101, 103 to 107;
  reserved "no_more";
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noChangeFieldIDsProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const changeFieldIDsProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4443;
  bool bar = 59;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noChangingFieldTypesProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const changingFieldTypesProto = `syntax = "proto3";
package test;

message Channel {
  int32 id = 1;
  bool name = 2;
  string description = 3;
  string foo = 4;
  repeated bool bar = 5;
}

message NextRequest {}
message PreviousRequest {}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

func TestParseOnReader(t *testing.T) {
	r := strings.NewReader(simpleProto)
	_, err := parse(r)
	assert.NoError(t, err)
}

func TestUsingReservedFields(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noUsingReservedFieldsProto)
	updLock := parseTestProto(t, usingReservedFieldsProto)

	warnings, ok := NoUsingReservedFields(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 3)

	warnings, ok = NoUsingReservedFields(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestChangingFieldTypes(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noChangingFieldTypesProto)
	updLock := parseTestProto(t, changingFieldTypesProto)

	warnings, ok := NoChangingFieldTypes(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 3)

	warnings, ok = NoChangingFieldTypes(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestRemovingReservedFields(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noRemoveReservedFieldsProto)
	updLock := parseTestProto(t, removeReservedFieldsProto)

	warnings, ok := NoRemovingReservedFields(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 5)

	warnings, ok = NoRemovingReservedFields(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestChangingFieldIDs(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noChangeFieldIDsProto)
	updLock := parseTestProto(t, changeFieldIDsProto)

	warnings, ok := NoChangingFieldIDs(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 2)

	warnings, ok = NoChangingFieldIDs(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func parseTestProto(t *testing.T, proto string) Protolock {
	r := strings.NewReader(proto)
	entry, err := parse(r)
	assert.NoError(t, err)
	return Protolock{
		Definitions: []Definition{
			{
				Filepath: "io.Reader (test)",
				Def:      entry,
			},
		},
	}
}