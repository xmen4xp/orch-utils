// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package strfmt

import (
	bsonprim "github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/strfmt/bson"
)

func init() {
	var id ObjectId
	// register this format in the default registry
	Default.Add("bsonobjectid", &id, IsBSONObjectID)
}

// IsBSONObjectID returns true when the string is a valid BSON.ObjectId
func IsBSONObjectID(str string) bool {
	_, err := bsonprim.ObjectIDFromHex(str)
	return err == nil
}

// ObjectId represents a BSON object ID (alias to go.mongodb.org/mongo-driver/bson/primitive.ObjectID)
//
// swagger:strfmt bsonobjectid
type ObjectId bsonprim.ObjectID

// NewObjectId creates a ObjectId from a Hex String
func NewObjectId(hex string) ObjectId {
	oid, err := bsonprim.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return ObjectId(oid)
}

// MarshalText turns this instance into text
func (id ObjectId) MarshalText() ([]byte, error) {
	oid := bsonprim.ObjectID(id)
	if oid == bsonprim.NilObjectID {
		return nil, nil
	}
	return []byte(oid.Hex()), nil
}

// UnmarshalText hydrates this instance from text
func (id *ObjectId) UnmarshalText(data []byte) error { // validation is performed later on
	if len(data) == 0 {
		*id = ObjectId(bsonprim.NilObjectID)
		return nil
	}
	oidstr := string(data)
	oid, err := bsonprim.ObjectIDFromHex(oidstr)
	if err != nil {
		return err
	}
	*id = ObjectId(oid)
	return nil
}

func (id ObjectId) String() string {
	return bsonprim.ObjectID(id).String()
}

// MarshalJSON returns the ObjectId as JSON
func (id ObjectId) MarshalJSON() ([]byte, error) {
	return bsonprim.ObjectID(id).MarshalJSON()
}

// UnmarshalJSON sets the ObjectId from JSON
func (id *ObjectId) UnmarshalJSON(data []byte) error {
	var obj bsonprim.ObjectID
	if err := obj.UnmarshalJSON(data); err != nil {
		return err
	}
	*id = ObjectId(obj)
	return nil
}

// DeepCopyInto copies the receiver and writes its value into out.
func (id *ObjectId) DeepCopyInto(out *ObjectId) {
	*out = *id
}

// DeepCopy copies the receiver into a new ObjectId.
func (id *ObjectId) DeepCopy() *ObjectId {
	if id == nil {
		return nil
	}
	out := new(ObjectId)
	id.DeepCopyInto(out)
	return out
}
