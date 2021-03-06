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

message NextRequest {
  reserved 3;
  reserved "a_map";
}

message PreviousRequest {
  reserved 4;
  reserved "no_use";
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
  }
}

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

  message A {
    int32 id = 1;
    string name = 2;
  }
}

message NextRequest {
  string name = 1;
  map<string, int32> a_map = 3;
}

message PreviousRequest {
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
    string no_use = 3;
    float32 thing = 4;
  }
}

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

message NextRequest {
  reserved 3;
  reserved "a_map";
}

message PreviousRequest {
  reserved 4;
  reserved "no_use";
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
  }
}

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

message NextRequest {
  map<string, int32> a_map = 3;  
}

message PreviousRequest {
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
  }
}

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

message NextRequest {
  map<string, int64> a_map = 1;
}

message PreviousRequest {
  reserved 4;
  reserved "no_use";
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
  }
}

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

message NextRequest {
  map<string, int64> a_map = 2;
}

message PreviousRequest {
  reserved 4;
  reserved "no_use";
  oneof test_oneof {
    int64 id = 11;
    bool is_active = 32;
  }
}

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

message NextRequest {
  string name = 1;
  map<string, int32> a_map = 3;
}

message PreviousRequest {
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
  }
}

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

message NextRequest {
  string name = 1;
  map<int64, bool> a_map = 3;
}

message PreviousRequest {
  oneof test_oneof {
    int32 id = 1;
    bool is_active = 2;
  }
}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noChangingFieldNamesProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {
  map<string, bool> a_map = 1;
}

message PreviousRequest {
  oneof test_oneof {
    string name = 4;
    bool is_active = 9;
  }
}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const changingFieldNamesProto = `syntax = "proto3";
package test;

message Channel {
  reserved "name", "foo";
  int64 channel_id = 1;
  string name_2 = 2;
  string description_3 = 3;
  string foo_baz = 4;
  bool bar = 5;
}

message NextRequest {
  map<string, bool> b_map = 1;
}

message PreviousRequest {
  oneof test_oneof {
    string name_2 = 4;
    bool is_active = 9;
  }
}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noRemovingServicesRPCsProto = `syntax = "proto3";
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

const removingServicesRPCsProto = `syntax = "proto3";
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
}
`

const noChangingRPCSignatureProto = `syntax = "proto3";
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

const changingRPCSignatureProto = `syntax = "proto3";
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
  rpc Next(NextRequest) returns (ChannelDifferent);
  rpc Previous(stream PreviousRequest) returns (stream Channel);
}
`

const noRemovingFieldsWithoutReserveProto = `syntax = "proto3";
package test;

message Channel {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string foo = 4;
  bool bar = 5;
}

message NextRequest {
  map<int32, bool> a_map = 1; 
}

message PreviousRequest {
  oneof test_oneof {
    int64 id = 1;
    bool is_active = 2;
  }
}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const removingFieldsWithoutReserveProto = `syntax = "proto3";
package test;

message Channel {
  reserved 5;
  int64 id = 1;
  string name_new = 2;
  string description = 3;
  string foo = 4;
}

message NextRequest {
  reserved 1;
}

message PreviousRequest {
  reserved 1;
}

service ChannelChanger {
  rpc Next(stream NextRequest) returns (Channel);
  rpc Previous(PreviousRequest) returns (stream Channel);
}
`

const noConflictSameNameNestedMessages = `syntax = "proto3";
package main;

message A {
    message I {
        int32 index = 1;
    }

    string id = 1;
    I i = 2;
}

message B {
    message I {
        reserved 2;
        int32 index = 1;
    }

    string id = 1;
    I i = 2;
}
`

const shouldConflictNestedMessage = `syntax = "proto3";
package main;

message A {
    message I {
        int32 index = 1;
    }

    string id = 1;
    I i = 2;
}

message B {
    message I {
        int32 index = 1;
        string name = 2;
    }

    string id = 1;
    I i = 2;
}
`

func TestParseOnReader(t *testing.T) {
	r := strings.NewReader(simpleProto)
	_, err := parse(r)
	assert.NoError(t, err)
}

func TestChangingRPCSignature(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noChangingRPCSignatureProto)
	updLock := parseTestProto(t, changingRPCSignatureProto)

	warnings, ok := NoChangingRPCSignature(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 3)

	warnings, ok = NoChangingRPCSignature(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestRemovingServiceRPCs(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noRemovingServicesRPCsProto)
	updLock := parseTestProto(t, removingServicesRPCsProto)

	warnings, ok := NoRemovingRPCs(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 2)

	warnings, ok = NoRemovingRPCs(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestChangingFieldNames(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noChangingFieldNamesProto)
	updLock := parseTestProto(t, changingFieldNamesProto)

	warnings, ok := NoChangingFieldNames(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 6)

	warnings, ok = NoChangingFieldNames(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestChangingFieldTypes(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noChangingFieldTypesProto)
	updLock := parseTestProto(t, changingFieldTypesProto)

	warnings, ok := NoChangingFieldTypes(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 6)

	warnings, ok = NoChangingFieldTypes(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestUsingReservedFields(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noUsingReservedFieldsProto)
	updLock := parseTestProto(t, usingReservedFieldsProto)

	warnings, ok := NoUsingReservedFields(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 7)

	warnings, ok = NoUsingReservedFields(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestRemovingReservedFields(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noRemoveReservedFieldsProto)
	updLock := parseTestProto(t, removeReservedFieldsProto)

	warnings, ok := NoRemovingReservedFields(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 9)

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
	assert.Len(t, warnings, 5)

	warnings, ok = NoChangingFieldIDs(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestRemovingFieldsWithoutReserve(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noRemovingFieldsWithoutReserveProto)
	updLock := parseTestProto(t, removingFieldsWithoutReserveProto)

	warnings, ok := NoRemovingFieldsWithoutReserve(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 6)

	warnings, ok = NoRemovingFieldsWithoutReserve(updLock, updLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestNoConflictSameNameNestedMessages(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noConflictSameNameNestedMessages)

	warnings, ok := NoUsingReservedFields(curLock, curLock)
	assert.True(t, ok)
	assert.Len(t, warnings, 0)
}

func TestShouldConflictReusingFieldsNestedMessages(t *testing.T) {
	SetDebug(true)
	curLock := parseTestProto(t, noConflictSameNameNestedMessages)
	updLock := parseTestProto(t, shouldConflictNestedMessage)

	warnings, ok := NoUsingReservedFields(curLock, updLock)
	assert.False(t, ok)
	assert.Len(t, warnings, 1)
}

func parseTestProto(t *testing.T, proto string) Protolock {
	r := strings.NewReader(proto)
	entry, err := parse(r)
	assert.NoError(t, err)
	return Protolock{
		Definitions: []Definition{
			{
				Filepath: protopath("testdata/no-test.proto"),
				Def:      entry,
			},
		},
	}
}
