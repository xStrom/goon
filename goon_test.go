/*
 * Copyright (c) 2013 Matt Jibson <matt.jibson@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package goon

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"appengine/memcache"
)

// *[]S, *[]*S, *[]I, []S, []*S, []I
const (
	ivTypePtrToSliceOfStructs = iota
	ivTypePtrToSliceOfPtrsToStruct
	ivTypePtrToSliceOfInterfaces
	ivTypeSliceOfStructs
	ivTypeSliceOfPtrsToStruct
	ivTypeSliceOfInterfaces
	ivTypeTotal
)

const (
	ivModeDatastore = iota
	ivModeMemcache
	ivModeMemcacheAndDatastore
	ivModeLocalcache
	ivModeLocalcacheAndMemcache
	ivModeLocalcacheAndDatastore
	ivModeLocalcacheAndMemcacheAndDatastore
	ivModeTotal
)

// Have a bunch of different supported types to detect any wild errors
type ivItem struct {
	Id        int64     `datastore:"-" goon:"id"`
	Int       int       `datastore:"int,noindex"`
	Int8      int8      `datastore:"int8,noindex"`
	Int16     int16     `datastore:"int16,noindex"`
	Int32     int32     `datastore:"int32,noindex"`
	Int64     int64     `datastore:"int64,noindex"`
	Float32   float32   `datastore:"float32,noindex"`
	Float64   float64   `datastore:"float64,noindex"`
	Bool      bool      `datastore:"bool,noindex"`
	String    string    `datastore:"string,noindex"`
	ByteSlice []byte    `datastore:"byte_slice,noindex"`
	Time      time.Time `datastore:"time,noindex"`
	Casual    string
	Key       *datastore.Key
	BlobKey   appengine.BlobKey
	Sub       ivItemSub
	Subs      []ivItemSubs
}

type ivItemSub struct {
	Data string `datastore:"data,noindex"`
	Ints []int  `datastore:"ints,noindex"`
}

type ivItemSubs struct {
	Data string `datastore:"data,noindex"`
}

func (ivi *ivItem) ForInterface() {}

type ivItemI interface {
	ForInterface()
}

var ivItems []ivItem

func initializeIvItems(c appengine.Context) {
	ivItems = []ivItem{
		{Id: 1, Int: 123, Int8: 77, Int16: 13001, Int32: 1234567890, Int64: 123456789012345,
			Float32: (float32(10) / float32(3)), Float64: (float64(10000000) / float64(9998)),
			Bool: true, String: "one", ByteSlice: []byte{0xDE, 0xAD},
			Time: time.Now().Truncate(time.Microsecond), Casual: "clothes",
			Key: datastore.NewKey(c, "Fruit", "Apple", 0, nil), BlobKey: "fake #1",
			Sub: ivItemSub{Data: "yay #1", Ints: []int{1, 2, 3}},
			Subs: []ivItemSubs{
				{Data: "sub #1.1"},
				{Data: "sub #1.2"},
				{Data: "sub #1.3"}}},
		{Id: 2, Int: 124, Int8: 78, Int16: 13002, Int32: 1234567891, Int64: 123456789012346,
			Float32: (float32(10) / float32(3)), Float64: (float64(10000000) / float64(9998)),
			Bool: true, String: "two", ByteSlice: []byte{0xBE, 0xEF},
			Time: time.Now().Truncate(time.Microsecond), Casual: "manners",
			Key: datastore.NewKey(c, "Fruit", "Banana", 0, nil), BlobKey: "fake #2",
			Sub: ivItemSub{Data: "yay #2", Ints: []int{4, 5, 6}},
			Subs: []ivItemSubs{
				{Data: "sub #2.1"},
				{Data: "sub #2.2"},
				{Data: "sub #2.3"}}},
		{Id: 3, Int: 125, Int8: 79, Int16: 13003, Int32: 1234567892, Int64: 123456789012347,
			Float32: (float32(10) / float32(3)), Float64: (float64(10000000) / float64(9998)),
			Bool: true, String: "tri", ByteSlice: []byte{0xF0, 0x0D},
			Time: time.Now().Truncate(time.Microsecond), Casual: "weather",
			Key: datastore.NewKey(c, "Fruit", "Cherry", 0, nil), BlobKey: "fake #3",
			Sub: ivItemSub{Data: "yay #3", Ints: []int{7, 8, 9}},
			Subs: []ivItemSubs{
				{Data: "sub #3.1"},
				{Data: "sub #3.2"},
				{Data: "sub #3.3"}}}}
}

func getInputVarietySrc(t *testing.T, ivType int, indices ...int) interface{} {
	if ivType >= ivTypeTotal {
		t.Fatalf("Invalid input variety type! %v >= %v", ivType, ivTypeTotal)
		return nil
	}

	var result interface{}

	switch ivType {
	case ivTypePtrToSliceOfStructs:
		s := []ivItem{}
		for _, index := range indices {
			ivItemCopy := ivItems[index]
			s = append(s, ivItemCopy)
		}
		result = &s
	case ivTypePtrToSliceOfPtrsToStruct:
		s := []*ivItem{}
		for _, index := range indices {
			ivItemCopy := ivItems[index]
			s = append(s, &ivItemCopy)
		}
		result = &s
	case ivTypePtrToSliceOfInterfaces:
		s := []ivItemI{}
		for _, index := range indices {
			ivItemCopy := ivItems[index]
			s = append(s, &ivItemCopy)
		}
		result = &s
	case ivTypeSliceOfStructs:
		s := []ivItem{}
		for _, index := range indices {
			ivItemCopy := ivItems[index]
			s = append(s, ivItemCopy)
		}
		result = s
	case ivTypeSliceOfPtrsToStruct:
		s := []*ivItem{}
		for _, index := range indices {
			ivItemCopy := ivItems[index]
			s = append(s, &ivItemCopy)
		}
		result = s
	case ivTypeSliceOfInterfaces:
		s := []ivItemI{}
		for _, index := range indices {
			ivItemCopy := ivItems[index]
			s = append(s, &ivItemCopy)
		}
		result = s
	}

	return result
}

func getInputVarietyDst(t *testing.T, ivType int) interface{} {
	if ivType >= ivTypeTotal {
		t.Fatalf("Invalid input variety type! %v >= %v", ivType, ivTypeTotal)
		return nil
	}

	var result interface{}

	switch ivType {
	case ivTypePtrToSliceOfStructs:
		result = &[]ivItem{{Id: ivItems[0].Id}, {Id: ivItems[1].Id}, {Id: ivItems[2].Id}}
	case ivTypePtrToSliceOfPtrsToStruct:
		result = &[]*ivItem{{Id: ivItems[0].Id}, {Id: ivItems[1].Id}, {Id: ivItems[2].Id}}
	case ivTypePtrToSliceOfInterfaces:
		result = &[]ivItemI{&ivItem{Id: ivItems[0].Id}, &ivItem{Id: ivItems[1].Id}, &ivItem{Id: ivItems[2].Id}}
	case ivTypeSliceOfStructs:
		result = []ivItem{{Id: ivItems[0].Id}, {Id: ivItems[1].Id}, {Id: ivItems[2].Id}}
	case ivTypeSliceOfPtrsToStruct:
		result = []*ivItem{{Id: ivItems[0].Id}, {Id: ivItems[1].Id}, {Id: ivItems[2].Id}}
	case ivTypeSliceOfInterfaces:
		result = []ivItemI{&ivItem{Id: ivItems[0].Id}, &ivItem{Id: ivItems[1].Id}, &ivItem{Id: ivItems[2].Id}}
	}

	return result
}

func getPrettyIVMode(ivMode int) string {
	result := "N/A"

	switch ivMode {
	case ivModeDatastore:
		result = "DS"
	case ivModeMemcache:
		result = "MC"
	case ivModeMemcacheAndDatastore:
		result = "DS+MC"
	case ivModeLocalcache:
		result = "LC"
	case ivModeLocalcacheAndMemcache:
		result = "MC+LC"
	case ivModeLocalcacheAndDatastore:
		result = "DS+LC"
	case ivModeLocalcacheAndMemcacheAndDatastore:
		result = "DS+MC+LC"
	}

	return result
}

func getPrettyIVType(ivType int) string {
	result := "N/A"

	switch ivType {
	case ivTypePtrToSliceOfStructs:
		result = "*[]S"
	case ivTypePtrToSliceOfPtrsToStruct:
		result = "*[]*S"
	case ivTypePtrToSliceOfInterfaces:
		result = "*[]I"
	case ivTypeSliceOfStructs:
		result = "[]S"
	case ivTypeSliceOfPtrsToStruct:
		result = "[]*S"
	case ivTypeSliceOfInterfaces:
		result = "[]I"
	}

	return result
}

func ivWipe(t *testing.T, g *Goon, prettyInfo string) {
	// Make sure the datastore is clear of any previous tests
	// TODO: Batch this once goon gets more convenient batch delete support
	for _, ivi := range ivItems {
		if err := g.Delete(g.Key(ivi)); err != nil {
			t.Errorf("%s > Unexpected error on delete - %v", prettyInfo, err)
		}
	}

	// Make sure the caches are clear, so any caching is done by our specific test
	g.FlushLocalCache()
	memcache.Flush(g.context)
}

func ivGetMulti(t *testing.T, g *Goon, ref, dst interface{}, prettyInfo string) error {
	// Get our data back and make sure it's correct
	if err := g.GetMulti(dst); err != nil {
		t.Errorf("%s > Unexpected error on GetMulti - %v", prettyInfo, err)
		return err
	} else {
		dstLen := reflect.Indirect(reflect.ValueOf(dst)).Len()
		refLen := reflect.Indirect(reflect.ValueOf(ref)).Len()

		if dstLen != refLen {
			t.Errorf("%s > Unexpected dst len (%v) doesn't match ref len (%v)", prettyInfo, dstLen, refLen)
		} else if !reflect.DeepEqual(ref, dst) {
			t.Errorf("%s > Expected - %v, %v, %v - got %v, %v, %v", prettyInfo,
				reflect.Indirect(reflect.ValueOf(ref)).Index(0).Interface(),
				reflect.Indirect(reflect.ValueOf(ref)).Index(1).Interface(),
				reflect.Indirect(reflect.ValueOf(ref)).Index(2).Interface(),
				reflect.Indirect(reflect.ValueOf(dst)).Index(0).Interface(),
				reflect.Indirect(reflect.ValueOf(dst)).Index(1).Interface(),
				reflect.Indirect(reflect.ValueOf(dst)).Index(2).Interface())
		}
	}

	return nil
}

func validateInputVariety(t *testing.T, g *Goon, srcType, dstType, mode int) {
	if mode >= ivModeTotal {
		t.Fatalf("Invalid input variety mode! %v >= %v", mode, ivModeTotal)
		return
	}

	// Generate a nice debug info string for clear logging
	prettyInfo := getPrettyIVType(srcType) + " " + getPrettyIVType(dstType) + " " + getPrettyIVMode(mode)

	// This function just gets the entities based on a predefined list, helper for cache population
	loadIVItem := func(indices ...int) {
		for _, index := range indices {
			ivi := &ivItem{Id: ivItems[index].Id}
			if err := g.Get(ivi); err != nil {
				t.Errorf("%s > Unexpected error on get - %v", prettyInfo, err)
			} else if !reflect.DeepEqual(ivItems[index], *ivi) {
				t.Errorf("%s > Expected - %v, got %v", prettyInfo, ivItems[index], *ivi)
			}
		}
	}

	// Start with a clean slate
	ivWipe(t, g, prettyInfo)

	// Generate test data with the specified types
	src := getInputVarietySrc(t, srcType, 0, 1, 2)
	ref := getInputVarietySrc(t, dstType, 0, 1, 2)
	dst := getInputVarietyDst(t, dstType)

	// Save our test data
	if _, err := g.PutMulti(src); err != nil {
		t.Errorf("%s > Unexpected error on PutMulti - %v", prettyInfo, err)
	}

	// Set the caches into proper state based on given mode
	switch mode {
	case ivModeDatastore:
		g.FlushLocalCache()
		memcache.Flush(g.context)
	case ivModeMemcache:
		loadIVItem(0, 1, 2) // Left in memcache
		g.FlushLocalCache()
	case ivModeMemcacheAndDatastore:
		loadIVItem(0, 1) // Left in memcache
		g.FlushLocalCache()
	case ivModeLocalcache:
		loadIVItem(0, 1, 2) // Left in local cache
	case ivModeLocalcacheAndMemcache:
		loadIVItem(0) // Left in memcache
		g.FlushLocalCache()
		loadIVItem(1, 2) // Left in local cache
	case ivModeLocalcacheAndDatastore:
		loadIVItem(0, 1) // Left in local cache
	case ivModeLocalcacheAndMemcacheAndDatastore:
		loadIVItem(0) // Left in memcache
		g.FlushLocalCache()
		loadIVItem(1) // Left in local cache
	}

	// Get our data back and make sure it's correct
	ivGetMulti(t, g, ref, dst, prettyInfo)
}

func validateInputVarietyTXNPut(t *testing.T, g *Goon, srcType, dstType, mode int) {
	if mode >= ivModeTotal {
		t.Fatalf("Invalid input variety mode! %v >= %v", mode, ivModeTotal)
		return
	}

	// The following modes are redundant with the current goon transaction implementation
	switch mode {
	case ivModeMemcache:
		return
	case ivModeMemcacheAndDatastore:
		return
	case ivModeLocalcacheAndMemcache:
		return
	case ivModeLocalcacheAndMemcacheAndDatastore:
		return
	}

	// Generate a nice debug info string for clear logging
	prettyInfo := getPrettyIVType(srcType) + " " + getPrettyIVType(dstType) + " " + getPrettyIVMode(mode) + " TXNPut"

	// Start with a clean slate
	ivWipe(t, g, prettyInfo)

	// Generate test data with the specified types
	src := getInputVarietySrc(t, srcType, 0, 1, 2)
	ref := getInputVarietySrc(t, dstType, 0, 1, 2)
	dst := getInputVarietyDst(t, dstType)

	// Save our test data
	if err := g.RunInTransaction(func(tg *Goon) error {
		_, err := tg.PutMulti(src)
		return err
	}, &datastore.TransactionOptions{XG: true}); err != nil {
		t.Errorf("%s > Unexpected error on PutMulti - %v", prettyInfo, err)
	}

	// Set the caches into proper state based on given mode
	switch mode {
	case ivModeDatastore:
		g.FlushLocalCache()
		memcache.Flush(g.context)
	case ivModeLocalcache:
		// Entities already in local cache
	case ivModeLocalcacheAndDatastore:
		g.FlushLocalCache()
		memcache.Flush(g.context)

		subSrc := getInputVarietySrc(t, srcType, 0)

		if err := g.RunInTransaction(func(tg *Goon) error {
			_, err := tg.PutMulti(subSrc)
			return err
		}, &datastore.TransactionOptions{XG: true}); err != nil {
			t.Errorf("%s > Unexpected error on PutMulti - %v", prettyInfo, err)
		}
	}

	// Get our data back and make sure it's correct
	ivGetMulti(t, g, ref, dst, prettyInfo)
}

func validateInputVarietyTXNGet(t *testing.T, g *Goon, srcType, dstType, mode int) {
	if mode >= ivModeTotal {
		t.Fatalf("Invalid input variety mode! %v >= %v", mode, ivModeTotal)
		return
	}

	// The following modes are redundant with the current goon transaction implementation
	switch mode {
	case ivModeMemcache:
		return
	case ivModeMemcacheAndDatastore:
		return
	case ivModeLocalcache:
		return
	case ivModeLocalcacheAndMemcache:
		return
	case ivModeLocalcacheAndDatastore:
		return
	case ivModeLocalcacheAndMemcacheAndDatastore:
		return
	}

	// Generate a nice debug info string for clear logging
	prettyInfo := getPrettyIVType(srcType) + " " + getPrettyIVType(dstType) + " " + getPrettyIVMode(mode) + " TXNGet"

	// Start with a clean slate
	ivWipe(t, g, prettyInfo)

	// Generate test data with the specified types
	src := getInputVarietySrc(t, srcType, 0, 1, 2)
	ref := getInputVarietySrc(t, dstType, 0, 1, 2)
	dst := getInputVarietyDst(t, dstType)

	// Save our test data
	if _, err := g.PutMulti(src); err != nil {
		t.Errorf("%s > Unexpected error on PutMulti - %v", prettyInfo, err)
	}

	// Set the caches into proper state based on given mode
	// TODO: Instead of clear, fill the caches with invalid data, because we're supposed to always fetch from the datastore
	switch mode {
	case ivModeDatastore:
		g.FlushLocalCache()
		memcache.Flush(g.context)
	}

	// Get our data back and make sure it's correct
	if err := g.RunInTransaction(func(tg *Goon) error {
		return ivGetMulti(t, tg, ref, dst, prettyInfo)
	}, &datastore.TransactionOptions{XG: true}); err != nil {
		t.Errorf("%s > Unexpected error on transaction - %v", prettyInfo, err)
	}
}

func TestInputVariety(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	g := FromContext(c)

	initializeIvItems(c)

	for srcType := 0; srcType < ivTypeTotal; srcType++ {
		for dstType := 0; dstType < ivTypeTotal; dstType++ {
			for mode := 0; mode < ivModeTotal; mode++ {
				validateInputVariety(t, g, srcType, dstType, mode)
				validateInputVarietyTXNPut(t, g, srcType, dstType, mode)
				validateInputVarietyTXNGet(t, g, srcType, dstType, mode)
			}
		}
	}
}

type MigrationA struct {
	_kind    string             `goon:"kind,Migration"`
	Id       int64              `datastore:"-" goon:"id"`
	Number   int32              `datastore:"number,noindex"`
	Word     string             `datastore:"word,noindex"`
	Car      string             `datastore:"car,noindex"`
	Sub      MigrationASub      `datastore:"sub,noindex"`
	Son      MigrationAPerson   `datastore:"son,noindex"`
	Daughter MigrationAPerson   `datastore:"daughter,noindex"`
	Parents  []MigrationAPerson `datastore:"parents,noindex"`
}

type MigrationASub struct {
	Data  string           `datastore:"data,noindex"`
	Noise []int            `datastore:"noise,noindex"`
	Sub   MigrationASubSub `datastore:"sub,noindex"`
}

type MigrationASubSub struct {
	Data string `datastore:"data,noindex"`
}

type MigrationAPerson struct {
	Name string `datastore:"name,noindex"`
	Age  int    `datastore:"age,noindex"`
}

type MigrationB struct {
	_kind          string             `goon:"kind,Migration"`
	Identification int64              `datastore:"-" goon:"id"`
	FancyNumber    int32              `datastore:"number,noindex"`
	Slang          string             `datastore:"word,noindex"`
	Cars           []string           `datastore:"car,noindex"`
	Animal         string             `datastore:"sub.data,noindex"`
	Music          []int              `datastore:"sub.noise,noindex"`
	Flower         string             `datastore:"sub.sub.data,noindex"`
	Sons           []MigrationAPerson `datastore:"son,noindex"`
	DaughterName   string             `datastore:"daughter.name,noindex"`
	DaughterAge    int                `datastore:"daughter.age,noindex"`
	OldFolks       []MigrationAPerson `datastore:"parents,noindex"`
}

func TestMigration(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	g := FromContext(c)

	// Create & save an entity with the original structure
	migA := &MigrationA{Id: 1, Number: 123, Word: "rabbit", Car: "BMW",
		Sub: MigrationASub{Data: "fox", Noise: []int{1, 2, 3}, Sub: MigrationASubSub{Data: "rose"}},
		Son: MigrationAPerson{Name: "John", Age: 5}, Daughter: MigrationAPerson{Name: "Nancy", Age: 6},
		Parents: []MigrationAPerson{{Name: "Sven", Age: 56}, {Name: "Sonya", Age: 49}}}
	if _, err := g.Put(migA); err != nil {
		t.Errorf("Unexpected error on Put: %v", err)
	}

	// Clear the local cache, because we want this data in memcache
	g.FlushLocalCache()

	// Get it back, so it's in the cache
	migA = &MigrationA{Id: 1}
	if err := g.Get(migA); err != nil {
		t.Errorf("Unexpected error on Get: %v", err)
	}

	// Clear the local cache, because it doesn't need to support migration
	g.FlushLocalCache()

	// Test whether memcache supports migration
	migB1 := &MigrationB{Identification: migA.Id}
	if err := g.Get(migB1); err != nil {
		t.Errorf("Unexpected error on Get: %v", err)
	} else if migA.Id != migB1.Identification {
		t.Errorf("Ids don't match: %v != %v", migA.Id, migB1.Identification)
	} else if migA.Number != migB1.FancyNumber {
		t.Errorf("Numbers don't match: %v != %v", migA.Number, migB1.FancyNumber)
	} else if migA.Word != migB1.Slang {
		t.Errorf("Words don't match: %v != %v", migA.Word, migB1.Slang)
	} else if len(migB1.Cars) != 1 {
		t.Errorf("Expected 1 car! Got: %v", len(migB1.Cars))
	} else if migA.Car != migB1.Cars[0] {
		t.Errorf("Cars don't match: %v != %v", migA.Car, migB1.Cars[0])
	} else if migA.Sub.Data != migB1.Animal {
		t.Errorf("Animal doesn't match: %v != %v", migA.Sub.Data, migB1.Animal)
	} else if !reflect.DeepEqual(migA.Sub.Noise, migB1.Music) {
		t.Errorf("Music doesn't match: %v != %v", migA.Sub.Noise, migB1.Music)
	} else if migA.Sub.Sub.Data != migB1.Flower {
		t.Errorf("Flower doesn't match: %v != %v", migA.Sub.Sub.Data, migB1.Flower)
	} else if len(migB1.Sons) != 1 {
		t.Errorf("Expected 1 son! Got: %v", len(migB1.Sons))
	} else if migA.Son.Name != migB1.Sons[0].Name {
		t.Errorf("Son names don't match: %v != %v", migA.Son.Name, migB1.Sons[0].Name)
	} else if migA.Son.Age != migB1.Sons[0].Age {
		t.Errorf("Son ages don't match: %v != %v", migA.Son.Age, migB1.Sons[0].Age)
	} else if migA.Daughter.Name != migB1.DaughterName {
		t.Errorf("Daughter names don't match: %v != %v", migA.Daughter.Name, migB1.DaughterName)
	} else if migA.Daughter.Age != migB1.DaughterAge {
		t.Errorf("Daughter ages don't match: %v != %v", migA.Daughter.Age, migB1.DaughterAge)
	} else if !reflect.DeepEqual(migA.Parents, migB1.OldFolks) {
		t.Errorf("Parents don't match: %v != %v", migA.Parents, migB1.OldFolks)
	}

	// Clear all the caches
	g.FlushLocalCache()
	memcache.Flush(c)

	// Test whether datastore supports migration
	migB2 := &MigrationB{Identification: migA.Id}
	if err := g.Get(migB2); err != nil {
		t.Errorf("Unexpected error on Get: %v", err)
	} else if migA.Id != migB2.Identification {
		t.Errorf("Ids don't match: %v != %v", migA.Id, migB2.Identification)
	} else if migA.Number != migB2.FancyNumber {
		t.Errorf("Numbers don't match: %v != %v", migA.Number, migB2.FancyNumber)
	} else if migA.Word != migB2.Slang {
		t.Errorf("Words don't match: %v != %v", migA.Word, migB2.Slang)
	} else if len(migB2.Cars) != 1 {
		t.Errorf("Expected 1 car! Got: %v", len(migB2.Cars))
	} else if migA.Car != migB2.Cars[0] {
		t.Errorf("Cars don't match: %v != %v", migA.Car, migB2.Cars[0])
	} else if migA.Sub.Data != migB2.Animal {
		t.Errorf("Animal doesn't match: %v != %v", migA.Sub.Data, migB2.Animal)
	} else if !reflect.DeepEqual(migA.Sub.Noise, migB2.Music) {
		t.Errorf("Music doesn't match: %v != %v", migA.Sub.Noise, migB2.Music)
	} else if migA.Sub.Sub.Data != migB2.Flower {
		t.Errorf("Flower doesn't match: %v != %v", migA.Sub.Sub.Data, migB2.Flower)
	} else if len(migB2.Sons) != 1 {
		t.Errorf("Expected 1 son! Got: %v", len(migB2.Sons))
	} else if migA.Son.Name != migB2.Sons[0].Name {
		t.Errorf("Sons don't match: %v != %v", migA.Son.Name, migB2.Sons[0].Name)
	} else if migA.Son.Age != migB2.Sons[0].Age {
		t.Errorf("Son ages don't match: %v != %v", migA.Son.Age, migB2.Sons[0].Age)
	} else if migA.Daughter.Name != migB2.DaughterName {
		t.Errorf("Daughters don't match: %v != %v", migA.Daughter.Name, migB2.DaughterName)
	} else if migA.Daughter.Age != migB2.DaughterAge {
		t.Errorf("Daughter ages don't match: %v != %v", migA.Daughter.Age, migB2.DaughterAge)
	} else if !reflect.DeepEqual(migA.Parents, migB2.OldFolks) {
		t.Errorf("Parents don't match: %v != %v", migA.Parents, migB2.OldFolks)
	}
}

func TestCaches(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	g := FromContext(c)

	// Put *struct{}
	phid := &HasId{Name: "cacheFail"}
	_, err = g.Put(phid)
	if err != nil {
		t.Errorf("Unexpected error on put - %v", err)
	}

	// fetch *struct{} from cache
	ghid := &HasId{Id: phid.Id}
	err = g.Get(ghid)
	if err != nil {
		t.Errorf("Unexpected error on get - %v", err)
	}
	if !reflect.DeepEqual(phid, ghid) {
		t.Errorf("Expected - %v, got %v", phid, ghid)
	}

	// fetch []struct{} from cache
	ghids := []HasId{{Id: phid.Id}}
	err = g.GetMulti(&ghids)
	if err != nil {
		t.Errorf("Unexpected error on get - %v", err)
	}
	if !reflect.DeepEqual(*phid, ghids[0]) {
		t.Errorf("Expected - %v, got %v", *phid, ghids[0])
	}

	// Now flush localcache and fetch them again
	g.FlushLocalCache()
	// fetch *struct{} from memcache
	ghid = &HasId{Id: phid.Id}
	err = g.Get(ghid)
	if err != nil {
		t.Errorf("Unexpected error on get - %v", err)
	}
	if !reflect.DeepEqual(phid, ghid) {
		t.Errorf("Expected - %v, got %v", phid, ghid)
	}

	g.FlushLocalCache()
	// fetch []struct{} from memcache
	ghids = []HasId{{Id: phid.Id}}
	err = g.GetMulti(&ghids)
	if err != nil {
		t.Errorf("Unexpected error on get - %v", err)
	}
	if !reflect.DeepEqual(*phid, ghids[0]) {
		t.Errorf("Expected - %v, got %v", *phid, ghids[0])
	}
}

func TestGoon(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	n := FromContext(c)

	// Don't want any of these tests to hit the timeout threshold on the devapp server
	MemcacheGetTimeout = time.Second
	MemcachePutTimeoutLarge = time.Second
	MemcachePutTimeoutSmall = time.Second

	// key tests
	noid := NoId{}
	if k, err := n.KeyError(noid); err == nil && !k.Incomplete() {
		t.Error("expected incomplete on noid")
	}
	if n.Key(noid) == nil {
		t.Error("expected to find a key")
	}

	var keyTests = []keyTest{
		{
			HasDefaultKind{},
			datastore.NewKey(c, "DefaultKind", "", 0, nil),
		},
		{
			HasId{Id: 1},
			datastore.NewKey(c, "HasId", "", 1, nil),
		},
		{
			HasKind{Id: 1, Kind: "OtherKind"},
			datastore.NewKey(c, "OtherKind", "", 1, nil),
		},

		{
			HasDefaultKind{Id: 1, Kind: "OtherKind"},
			datastore.NewKey(c, "OtherKind", "", 1, nil),
		},
		{
			HasDefaultKind{Id: 1},
			datastore.NewKey(c, "DefaultKind", "", 1, nil),
		},
		{
			HasString{Id: "new"},
			datastore.NewKey(c, "HasString", "new", 0, nil),
		},
	}

	for _, kt := range keyTests {
		if k, err := n.KeyError(kt.obj); err != nil {
			t.Errorf("error:", err)
		} else if !k.Equal(kt.key) {
			t.Errorf("keys not equal")
		}
	}

	if _, err := n.KeyError(TwoId{IntId: 1, StringId: "1"}); err == nil {
		t.Errorf("expected key error")
	}

	// datastore tests
	keys, _ := datastore.NewQuery("HasId").KeysOnly().GetAll(c, nil)
	datastore.DeleteMulti(c, keys)
	memcache.Flush(c)
	if err := n.Get(&HasId{Id: 0}); err == nil {
		t.Errorf("ds: expected error, we're fetching from the datastore on an incomplete key!")
	}
	if err := n.Get(&HasId{Id: 1}); err != datastore.ErrNoSuchEntity {
		t.Errorf("ds: expected no such entity")
	}
	// run twice to make sure autocaching works correctly
	if err := n.Get(&HasId{Id: 1}); err != datastore.ErrNoSuchEntity {
		t.Errorf("ds: expected no such entity")
	}
	es := []*HasId{
		{Id: 1, Name: "one"},
		{Id: 2, Name: "two"},
	}
	var esk []*datastore.Key
	for _, e := range es {
		esk = append(esk, n.Key(e))
	}
	nes := []*HasId{
		{Id: 1},
		{Id: 2},
	}
	if err := n.GetMulti(es); err == nil {
		t.Errorf("ds: expected error")
	} else if !NotFound(err, 0) {
		t.Errorf("ds: not found error 0")
	} else if !NotFound(err, 1) {
		t.Errorf("ds: not found error 1")
	} else if NotFound(err, 2) {
		t.Errorf("ds: not found error 2")
	}

	if keys, err := n.PutMulti(es); err != nil {
		t.Errorf("put: unexpected error")
	} else if len(keys) != len(esk) {
		t.Errorf("put: got unexpected number of keys")
	} else {
		for i, k := range keys {
			if !k.Equal(esk[i]) {
				t.Errorf("put: got unexpected keys")
			}
		}
	}
	if err := n.GetMulti(nes); err != nil {
		t.Errorf("put: unexpected error")
	} else if *es[0] != *nes[0] || *es[1] != *nes[1] {
		t.Errorf("put: bad results")
	} else {
		nesk0 := n.Key(nes[0])
		if !nesk0.Equal(datastore.NewKey(c, "HasId", "", 1, nil)) {
			t.Errorf("put: bad key")
		}
		nesk1 := n.Key(nes[1])
		if !nesk1.Equal(datastore.NewKey(c, "HasId", "", 2, nil)) {
			t.Errorf("put: bad key")
		}
	}
	if _, err := n.Put(HasId{Id: 3}); err == nil {
		t.Errorf("put: expected error")
	}
	// force partial fetch from memcache and then datastore
	memcache.Flush(c)
	if err := n.Get(nes[0]); err != nil {
		t.Errorf("get: unexpected error")
	}
	if err := n.GetMulti(nes); err != nil {
		t.Errorf("get: unexpected error")
	}

	// put a HasId resource, then test pulling it from memory, memcache, and datastore
	hi := &HasId{Name: "hasid"} // no id given, should be automatically created by the datastore
	if _, err := n.Put(hi); err != nil {
		t.Errorf("put: unexpected error - %v", err)
	}
	if n.Key(hi) == nil {
		t.Errorf("key should not be nil")
	} else if n.Key(hi).Incomplete() {
		t.Errorf("key should not be incomplete")
	}

	hi2 := &HasId{Id: hi.Id}
	if err := n.Get(hi2); err != nil {
		t.Errorf("get: unexpected error - %v", err)
	}
	if hi2.Name != hi.Name {
		t.Errorf("Could not fetch HasId object from memory - %#v != %#v, memory=%#v", hi, hi2, n.cache[memkey(n.Key(hi2))])
	}

	hi3 := &HasId{Id: hi.Id}
	delete(n.cache, memkey(n.Key(hi)))
	if err := n.Get(hi3); err != nil {
		t.Errorf("get: unexpected error - %v", err)
	}
	if hi3.Name != hi.Name {
		t.Errorf("Could not fetch HasId object from memory - %#v != %#v", hi, hi3)
	}

	hi4 := &HasId{Id: hi.Id}
	delete(n.cache, memkey(n.Key(hi4)))
	if memcache.Flush(n.context) != nil {
		t.Errorf("Unable to flush memcache")
	}
	if err := n.Get(hi4); err != nil {
		t.Errorf("get: unexpected error - %v", err)
	}
	if hi4.Name != hi.Name {
		t.Errorf("Could not fetch HasId object from datastore- %#v != %#v", hi, hi4)
	}

	// Now do the opposite also using hi
	// Test pulling from local cache and memcache when datastore result is different
	// Note that this shouldn't happen with real goon usage,
	//   but this tests that goon isn't still pulling from the datastore (or memcache) unnecessarily
	// hi in datastore Name = hasid
	hiPull := &HasId{Id: hi.Id}
	n.cacheLock.Lock()
	n.cache[memkey(n.Key(hi))].(*HasId).Name = "changedincache"
	n.cacheLock.Unlock()
	if err := n.Get(hiPull); err != nil {
		t.Errorf("get: unexpected error - %v", err)
	}
	if hiPull.Name != "changedincache" {
		t.Errorf("hiPull.Name should be 'changedincache' but got %s", hiPull.Name)
	}

	hiPush := &HasId{Id: hi.Id, Name: "changedinmemcache"}
	n.putMemcache([]interface{}{hiPush})
	n.cacheLock.Lock()
	delete(n.cache, memkey(n.Key(hi)))
	n.cacheLock.Unlock()

	hiPull = &HasId{Id: hi.Id}
	if err := n.Get(hiPull); err != nil {
		t.Errorf("get: unexpected error - %v", err)
	}
	if hiPull.Name != "changedinmemcache" {
		t.Errorf("hiPull.Name should be 'changedinmemcache' but got %s", hiPull.Name)
	}

	// Since the datastore can't assign a key to a String ID, test to make sure goon stops it from happening
	hasString := new(HasString)
	_, err = n.Put(hasString)
	if err == nil {
		t.Errorf("Cannot put an incomplete string Id object as the datastore will populate an int64 id instead- %v", hasString)
	}
	hasString.Id = "hello"
	_, err = n.Put(hasString)
	if err != nil {
		t.Errorf("Error putting hasString object - %v", hasString)
	}

	// Test queries!

	// Test that zero result queries work properly
	qiZRes := []QueryItem{}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem"), &qiZRes); err != nil {
		t.Errorf("GetAll Zero: unexpected error: %v", err)
	} else if len(dskeys) != 0 {
		t.Errorf("GetAll Zero: expected 0 keys, got %v", len(dskeys))
	}

	// Create some entities that we will query for
	if _, err := n.PutMulti([]*QueryItem{{Id: 1, Data: "one"}, {Id: 2, Data: "two"}}); err != nil {
		t.Errorf("PutMulti: unexpected error: %v", err)
	}

	// Sleep a bit to wait for the HRD emulation to get out of our way
	time.Sleep(1000 * time.Millisecond)

	// Clear the local memory cache, because we want to test it being filled correctly by GetAll
	n.FlushLocalCache()

	// Get the entity using a slice of structs
	qiSRes := []QueryItem{}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "one"), &qiSRes); err != nil {
		t.Errorf("GetAll SoS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll SoS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 1 {
		t.Errorf("GetAll SoS: expected key IntID to be 1, got %v", dskeys[0].IntID())
	} else if len(qiSRes) != 1 {
		t.Errorf("GetAll SoS: expected 1 result, got %v", len(qiSRes))
	} else if qiSRes[0].Id != 1 {
		t.Errorf("GetAll SoS: expected entity id to be 1, got %v", qiSRes[0].Id)
	} else if qiSRes[0].Data != "one" {
		t.Errorf("GetAll SoS: expected entity data to be 'one', got '%v'", qiSRes[0].Data)
	}

	// Get the entity using normal Get to test local cache (provided the local cache actually got saved)
	qiS := &QueryItem{Id: 1}
	if err := n.Get(qiS); err != nil {
		t.Errorf("Get SoS: unexpected error: %v", err)
	} else if qiS.Id != 1 {
		t.Errorf("Get SoS: expected entity id to be 1, got %v", qiS.Id)
	} else if qiS.Data != "one" {
		t.Errorf("Get SoS: expected entity data to be 'one', got '%v'", qiS.Data)
	}

	// Clear the local memory cache, because we want to test it being filled correctly by GetAll
	n.FlushLocalCache()

	// Get the entity using a slice of pointers to struct
	qiPRes := []*QueryItem{}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "one"), &qiPRes); err != nil {
		t.Errorf("GetAll SoPtS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll SoPtS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 1 {
		t.Errorf("GetAll SoPtS: expected key IntID to be 1, got %v", dskeys[0].IntID())
	} else if len(qiPRes) != 1 {
		t.Errorf("GetAll SoPtS: expected 1 result, got %v", len(qiPRes))
	} else if qiPRes[0].Id != 1 {
		t.Errorf("GetAll SoPtS: expected entity id to be 1, got %v", qiPRes[0].Id)
	} else if qiPRes[0].Data != "one" {
		t.Errorf("GetAll SoPtS: expected entity data to be 'one', got '%v'", qiPRes[0].Data)
	}

	// Get the entity using normal Get to test local cache (provided the local cache actually got saved)
	qiP := &QueryItem{Id: 1}
	if err := n.Get(qiP); err != nil {
		t.Errorf("Get SoPtS: unexpected error: %v", err)
	} else if qiP.Id != 1 {
		t.Errorf("Get SoPtS: expected entity id to be 1, got %v", qiP.Id)
	} else if qiP.Data != "one" {
		t.Errorf("Get SoPtS: expected entity data to be 'one', got '%v'", qiP.Data)
	}

	// Clear the local memory cache, because we want to test it being filled correctly by Next
	n.FlushLocalCache()

	// Get the entity using an iterator
	qiIt := n.Run(datastore.NewQuery("QueryItem").Filter("data=", "one"))

	qiItRes := &QueryItem{}
	if dskey, err := qiIt.Next(qiItRes); err != nil {
		t.Errorf("Next: unexpected error: %v", err)
	} else if dskey.IntID() != 1 {
		t.Errorf("Next: expected key IntID to be 1, got %v", dskey.IntID())
	} else if qiItRes.Id != 1 {
		t.Errorf("Next: expected entity id to be 1, got %v", qiItRes.Id)
	} else if qiItRes.Data != "one" {
		t.Errorf("Next: expected entity data to be 'one', got '%v'", qiItRes.Data)
	}

	// Make sure the iterator ends correctly
	if _, err := qiIt.Next(&QueryItem{}); err != datastore.Done {
		t.Errorf("Next: expected iterator to end with the error datastore.Done, got %v", err)
	}

	// Get the entity using normal Get to test local cache (provided the local cache actually got saved)
	qiI := &QueryItem{Id: 1}
	if err := n.Get(qiI); err != nil {
		t.Errorf("Get Iterator: unexpected error: %v", err)
	} else if qiI.Id != 1 {
		t.Errorf("Get Iterator: expected entity id to be 1, got %v", qiI.Id)
	} else if qiI.Data != "one" {
		t.Errorf("Get Iterator: expected entity data to be 'one', got '%v'", qiI.Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly when supplying a non-zero slice
	n.FlushLocalCache()

	// Get the entity using a non-zero slice of structs
	qiNZSRes := []QueryItem{{Id: 1, Data: "invalid cache"}}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "two"), &qiNZSRes); err != nil {
		t.Errorf("GetAll NZSoS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll NZSoS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 2 {
		t.Errorf("GetAll NZSoS: expected key IntID to be 2, got %v", dskeys[0].IntID())
	} else if len(qiNZSRes) != 2 {
		t.Errorf("GetAll NZSoS: expected slice len to be 2, got %v", len(qiNZSRes))
	} else if qiNZSRes[0].Id != 1 {
		t.Errorf("GetAll NZSoS: expected entity id to be 1, got %v", qiNZSRes[0].Id)
	} else if qiNZSRes[0].Data != "invalid cache" {
		t.Errorf("GetAll NZSoS: expected entity data to be 'invalid cache', got '%v'", qiNZSRes[0].Data)
	} else if qiNZSRes[1].Id != 2 {
		t.Errorf("GetAll NZSoS: expected entity id to be 2, got %v", qiNZSRes[1].Id)
	} else if qiNZSRes[1].Data != "two" {
		t.Errorf("GetAll NZSoS: expected entity data to be 'two', got '%v'", qiNZSRes[1].Data)
	}

	// Get the entities using normal GetMulti to test local cache
	qiNZSs := []QueryItem{{Id: 1}, {Id: 2}}
	if err := n.GetMulti(qiNZSs); err != nil {
		t.Errorf("GetMulti NZSoS: unexpected error: %v", err)
	} else if len(qiNZSs) != 2 {
		t.Errorf("GetMulti NZSoS: expected slice len to be 2, got %v", len(qiNZSs))
	} else if qiNZSs[0].Id != 1 {
		t.Errorf("GetMulti NZSoS: expected entity id to be 1, got %v", qiNZSs[0].Id)
	} else if qiNZSs[0].Data != "one" {
		t.Errorf("GetMulti NZSoS: expected entity data to be 'one', got '%v'", qiNZSs[0].Data)
	} else if qiNZSs[1].Id != 2 {
		t.Errorf("GetMulti NZSoS: expected entity id to be 2, got %v", qiNZSs[1].Id)
	} else if qiNZSs[1].Data != "two" {
		t.Errorf("GetMulti NZSoS: expected entity data to be 'two', got '%v'", qiNZSs[1].Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly when supplying a non-zero slice
	n.FlushLocalCache()

	// Get the entity using a non-zero slice of pointers to struct
	qiNZPRes := []*QueryItem{{Id: 1, Data: "invalid cache"}}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "two"), &qiNZPRes); err != nil {
		t.Errorf("GetAll NZSoPtS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll NZSoPtS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 2 {
		t.Errorf("GetAll NZSoPtS: expected key IntID to be 2, got %v", dskeys[0].IntID())
	} else if len(qiNZPRes) != 2 {
		t.Errorf("GetAll NZSoPtS: expected slice len to be 2, got %v", len(qiNZPRes))
	} else if qiNZPRes[0].Id != 1 {
		t.Errorf("GetAll NZSoPtS: expected entity id to be 1, got %v", qiNZPRes[0].Id)
	} else if qiNZPRes[0].Data != "invalid cache" {
		t.Errorf("GetAll NZSoPtS: expected entity data to be 'invalid cache', got '%v'", qiNZPRes[0].Data)
	} else if qiNZPRes[1].Id != 2 {
		t.Errorf("GetAll NZSoPtS: expected entity id to be 2, got %v", qiNZPRes[1].Id)
	} else if qiNZPRes[1].Data != "two" {
		t.Errorf("GetAll NZSoPtS: expected entity data to be 'two', got '%v'", qiNZPRes[1].Data)
	}

	// Get the entities using normal GetMulti to test local cache
	qiNZPs := []*QueryItem{{Id: 1}, {Id: 2}}
	if err := n.GetMulti(qiNZPs); err != nil {
		t.Errorf("GetMulti NZSoPtS: unexpected error: %v", err)
	} else if len(qiNZPs) != 2 {
		t.Errorf("GetMulti NZSoPtS: expected slice len to be 2, got %v", len(qiNZPs))
	} else if qiNZPs[0].Id != 1 {
		t.Errorf("GetMulti NZSoPtS: expected entity id to be 1, got %v", qiNZPs[0].Id)
	} else if qiNZPs[0].Data != "one" {
		t.Errorf("GetMulti NZSoPtS: expected entity data to be 'one', got '%v'", qiNZPs[0].Data)
	} else if qiNZPs[1].Id != 2 {
		t.Errorf("GetMulti NZSoPtS: expected entity id to be 2, got %v", qiNZPs[1].Id)
	} else if qiNZPs[1].Data != "two" {
		t.Errorf("GetMulti NZSoPtS: expected entity data to be 'two', got '%v'", qiNZPs[1].Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly by a keys-only query
	n.FlushLocalCache()

	// Test the simplest keys-only query
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "one").KeysOnly(), nil); err != nil {
		t.Errorf("GetAll KeysOnly: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll KeysOnly: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 1 {
		t.Errorf("GetAll KeysOnly: expected key IntID to be 1, got %v", dskeys[0].IntID())
	}

	// Get the entity using normal Get to test that the local cache wasn't filled with incomplete data
	qiKO := &QueryItem{Id: 1}
	if err := n.Get(qiKO); err != nil {
		t.Errorf("Get KeysOnly: unexpected error: %v", err)
	} else if qiKO.Id != 1 {
		t.Errorf("Get KeysOnly: expected entity id to be 1, got %v", qiKO.Id)
	} else if qiKO.Data != "one" {
		t.Errorf("Get KeysOnly: expected entity data to be 'one', got '%v'", qiKO.Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly by a keys-only query
	n.FlushLocalCache()

	// Test the keys-only query with slice of structs
	qiKOSRes := []QueryItem{}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "one").KeysOnly(), &qiKOSRes); err != nil {
		t.Errorf("GetAll KeysOnly SoS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll KeysOnly SoS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 1 {
		t.Errorf("GetAll KeysOnly SoS: expected key IntID to be 1, got %v", dskeys[0].IntID())
	} else if len(qiKOSRes) != 1 {
		t.Errorf("GetAll KeysOnly SoS: expected 1 result, got %v", len(qiKOSRes))
	} else if k := reflect.TypeOf(qiKOSRes[0]).Kind(); k != reflect.Struct {
		t.Errorf("GetAll KeysOnly SoS: expected struct, got %v", k)
	} else if qiKOSRes[0].Id != 1 {
		t.Errorf("GetAll KeysOnly SoS: expected entity id to be 1, got %v", qiKOSRes[0].Id)
	} else if qiKOSRes[0].Data != "" {
		t.Errorf("GetAll KeysOnly SoS: expected entity data to be empty, got '%v'", qiKOSRes[0].Data)
	}

	// Get the entity using normal Get to test that the local cache wasn't filled with incomplete data
	if err := n.GetMulti(qiKOSRes); err != nil {
		t.Errorf("Get KeysOnly SoS: unexpected error: %v", err)
	} else if qiKOSRes[0].Id != 1 {
		t.Errorf("Get KeysOnly SoS: expected entity id to be 1, got %v", qiKOSRes[0].Id)
	} else if qiKOSRes[0].Data != "one" {
		t.Errorf("Get KeysOnly SoS: expected entity data to be 'one', got '%v'", qiKOSRes[0].Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly by a keys-only query
	n.FlushLocalCache()

	// Test the keys-only query with slice of pointers to struct
	qiKOPRes := []*QueryItem{}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "one").KeysOnly(), &qiKOPRes); err != nil {
		t.Errorf("GetAll KeysOnly SoPtS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll KeysOnly SoPtS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 1 {
		t.Errorf("GetAll KeysOnly SoPtS: expected key IntID to be 1, got %v", dskeys[0].IntID())
	} else if len(qiKOPRes) != 1 {
		t.Errorf("GetAll KeysOnly SoPtS: expected 1 result, got %v", len(qiKOPRes))
	} else if k := reflect.TypeOf(qiKOPRes[0]).Kind(); k != reflect.Ptr {
		t.Errorf("GetAll KeysOnly SoPtS: expected pointer, got %v", k)
	} else if qiKOPRes[0].Id != 1 {
		t.Errorf("GetAll KeysOnly SoPtS: expected entity id to be 1, got %v", qiKOPRes[0].Id)
	} else if qiKOPRes[0].Data != "" {
		t.Errorf("GetAll KeysOnly SoPtS: expected entity data to be empty, got '%v'", qiKOPRes[0].Data)
	}

	// Get the entity using normal Get to test that the local cache wasn't filled with incomplete data
	if err := n.GetMulti(qiKOPRes); err != nil {
		t.Errorf("Get KeysOnly SoPtS: unexpected error: %v", err)
	} else if qiKOPRes[0].Id != 1 {
		t.Errorf("Get KeysOnly SoPtS: expected entity id to be 1, got %v", qiKOPRes[0].Id)
	} else if qiKOPRes[0].Data != "one" {
		t.Errorf("Get KeysOnly SoPtS: expected entity data to be 'one', got '%v'", qiKOPRes[0].Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly when supplying a non-zero slice
	n.FlushLocalCache()

	// Test the keys-only query with non-zero slice of structs
	qiKONZSRes := []QueryItem{{Id: 1, Data: "invalid cache"}}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "two").KeysOnly(), &qiKONZSRes); err != nil {
		t.Errorf("GetAll KeysOnly NZSoS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll KeysOnly NZSoS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 2 {
		t.Errorf("GetAll KeysOnly NZSoS: expected key IntID to be 2, got %v", dskeys[0].IntID())
	} else if len(qiKONZSRes) != 2 {
		t.Errorf("GetAll KeysOnly NZSoS: expected slice len to be 2, got %v", len(qiKONZSRes))
	} else if qiKONZSRes[0].Id != 1 {
		t.Errorf("GetAll KeysOnly NZSoS: expected entity id to be 1, got %v", qiKONZSRes[0].Id)
	} else if qiKONZSRes[0].Data != "invalid cache" {
		t.Errorf("GetAll KeysOnly NZSoS: expected entity data to be 'invalid cache', got '%v'", qiKONZSRes[0].Data)
	} else if k := reflect.TypeOf(qiKONZSRes[1]).Kind(); k != reflect.Struct {
		t.Errorf("GetAll KeysOnly NZSoS: expected struct, got %v", k)
	} else if qiKONZSRes[1].Id != 2 {
		t.Errorf("GetAll KeysOnly NZSoS: expected entity id to be 2, got %v", qiKONZSRes[1].Id)
	} else if qiKONZSRes[1].Data != "" {
		t.Errorf("GetAll KeysOnly NZSoS: expected entity data to be empty, got '%v'", qiKONZSRes[1].Data)
	}

	// Get the entities using normal GetMulti to test local cache
	if err := n.GetMulti(qiKONZSRes); err != nil {
		t.Errorf("GetMulti NZSoS: unexpected error: %v", err)
	} else if len(qiKONZSRes) != 2 {
		t.Errorf("GetMulti NZSoS: expected slice len to be 2, got %v", len(qiKONZSRes))
	} else if qiKONZSRes[0].Id != 1 {
		t.Errorf("GetMulti NZSoS: expected entity id to be 1, got %v", qiKONZSRes[0].Id)
	} else if qiKONZSRes[0].Data != "one" {
		t.Errorf("GetMulti NZSoS: expected entity data to be 'one', got '%v'", qiKONZSRes[0].Data)
	} else if qiKONZSRes[1].Id != 2 {
		t.Errorf("GetMulti NZSoS: expected entity id to be 2, got %v", qiKONZSRes[1].Id)
	} else if qiKONZSRes[1].Data != "two" {
		t.Errorf("GetMulti NZSoS: expected entity data to be 'two', got '%v'", qiKONZSRes[1].Data)
	}

	// Clear the local memory cache, because we want to test it not being filled incorrectly when supplying a non-zero slice
	n.FlushLocalCache()

	// Test the keys-only query with non-zero slice of pointers to struct
	qiKONZPRes := []*QueryItem{{Id: 1, Data: "invalid cache"}}
	if dskeys, err := n.GetAll(datastore.NewQuery("QueryItem").Filter("data=", "two").KeysOnly(), &qiKONZPRes); err != nil {
		t.Errorf("GetAll KeysOnly NZSoPtS: unexpected error: %v", err)
	} else if len(dskeys) != 1 {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected 1 key, got %v", len(dskeys))
	} else if dskeys[0].IntID() != 2 {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected key IntID to be 2, got %v", dskeys[0].IntID())
	} else if len(qiKONZPRes) != 2 {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected slice len to be 2, got %v", len(qiKONZPRes))
	} else if qiKONZPRes[0].Id != 1 {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected entity id to be 1, got %v", qiKONZPRes[0].Id)
	} else if qiKONZPRes[0].Data != "invalid cache" {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected entity data to be 'invalid cache', got '%v'", qiKONZPRes[0].Data)
	} else if k := reflect.TypeOf(qiKONZPRes[1]).Kind(); k != reflect.Ptr {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected pointer, got %v", k)
	} else if qiKONZPRes[1].Id != 2 {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected entity id to be 2, got %v", qiKONZPRes[1].Id)
	} else if qiKONZPRes[1].Data != "" {
		t.Errorf("GetAll KeysOnly NZSoPtS: expected entity data to be empty, got '%v'", qiKONZPRes[1].Data)
	}

	// Get the entities using normal GetMulti to test local cache
	if err := n.GetMulti(qiKONZPRes); err != nil {
		t.Errorf("GetMulti NZSoPtS: unexpected error: %v", err)
	} else if len(qiKONZPRes) != 2 {
		t.Errorf("GetMulti NZSoPtS: expected slice len to be 2, got %v", len(qiKONZPRes))
	} else if qiKONZPRes[0].Id != 1 {
		t.Errorf("GetMulti NZSoPtS: expected entity id to be 1, got %v", qiKONZPRes[0].Id)
	} else if qiKONZPRes[0].Data != "one" {
		t.Errorf("GetMulti NZSoPtS: expected entity data to be 'one', got '%v'", qiKONZPRes[0].Data)
	} else if qiKONZPRes[1].Id != 2 {
		t.Errorf("GetMulti NZSoPtS: expected entity id to be 2, got %v", qiKONZPRes[1].Id)
	} else if qiKONZPRes[1].Data != "two" {
		t.Errorf("GetMulti NZSoPtS: expected entity data to be 'two', got '%v'", qiKONZPRes[1].Data)
	}
}

type keyTest struct {
	obj interface{}
	key *datastore.Key
}

type NoId struct {
}

type HasId struct {
	Id   int64 `datastore:"-" goon:"id"`
	Name string
}

type HasKind struct {
	Id   int64  `datastore:"-" goon:"id"`
	Kind string `datastore:"-" goon:"kind"`
	Name string
}

type HasDefaultKind struct {
	Id   int64  `datastore:"-" goon:"id"`
	Kind string `datastore:"-" goon:"kind,DefaultKind"`
	Name string
}

type QueryItem struct {
	Id   int64  `datastore:"-" goon:"id"`
	Data string `datastore:"data"`
}

type HasString struct {
	Id string `datastore:"-" goon:"id"`
}

type TwoId struct {
	IntId    int64  `goon:"id"`
	StringId string `goon:"id"`
}

type PutGet struct {
	ID    int64 `datastore:"-" goon:"id"`
	Value int32
}

// Commenting out for issue https://code.google.com/p/googleappengine/issues/detail?id=10493
//func TestMemcachePutTimeout(t *testing.T) {
//	c, err := aetest.NewContext(nil)
//	if err != nil {
//		t.Fatalf("Could not start aetest - %v", err)
//	}
//	defer c.Close()
//	g := FromContext(c)
//	MemcachePutTimeoutSmall = time.Second
//	// put a HasId resource, then test pulling it from memory, memcache, and datastore
//	hi := &HasId{Name: "hasid"} // no id given, should be automatically created by the datastore
//	if _, err := g.Put(hi); err != nil {
//		t.Errorf("put: unexpected error - %v", err)
//	}

//	MemcachePutTimeoutSmall = 0
//	MemcacheGetTimeout = 0
//	if err := g.putMemcache([]interface{}{hi}); !appengine.IsTimeoutError(err) {
//		t.Errorf("Request should timeout - err = %v", err)
//	}
//	MemcachePutTimeoutSmall = time.Second
//	MemcachePutTimeoutThreshold = 0
//	MemcachePutTimeoutLarge = 0
//	if err := g.putMemcache([]interface{}{hi}); !appengine.IsTimeoutError(err) {
//		t.Errorf("Request should timeout - err = %v", err)
//	}

//	MemcachePutTimeoutLarge = time.Second
//	if err := g.putMemcache([]interface{}{hi}); err != nil {
//		t.Errorf("putMemcache: unexpected error - %v", err)
//	}

//	g.FlushLocalCache()
//	memcache.Flush(c)
//	// time out Get
//	MemcacheGetTimeout = 0
//	// time out Put too
//	MemcachePutTimeoutSmall = 0
//	MemcachePutTimeoutThreshold = 0
//	MemcachePutTimeoutLarge = 0
//	hiResult := &HasId{Id: hi.Id}
//	if err := g.Get(hiResult); err != nil {
//		t.Errorf("Request should not timeout cause we'll fetch from the datastore but got error  %v", err)
//		// Put timing out should also error, but it won't be returned here, just logged
//	}
//	if !reflect.DeepEqual(hi, hiResult) {
//		t.Errorf("Fetched object isn't accurate - want %v, fetched %v", hi, hiResult)
//	}

//	hiResult = &HasId{Id: hi.Id}
//	g.FlushLocalCache()
//	MemcacheGetTimeout = time.Second
//	if err := g.Get(hiResult); err != nil {
//		t.Errorf("Request should not timeout cause we'll fetch from memcache successfully but got error %v", err)
//	}
//	if !reflect.DeepEqual(hi, hiResult) {
//		t.Errorf("Fetched object isn't accurate - want %v, fetched %v", hi, hiResult)
//	}
//}

// This test won't fail but if run with -race flag, it will show known race conditions
// Using multiple goroutines per http request is recommended here:
// http://talks.golang.org/2013/highperf.slide#22
func TestRace(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	g := FromContext(c)

	var hasIdSlice []*HasId
	for x := 1; x <= 4000; x++ {
		hasIdSlice = append(hasIdSlice, &HasId{Id: int64(x), Name: "Race"})
	}
	_, err = g.PutMulti(hasIdSlice)
	if err != nil {
		t.Fatalf("Could not put Race entities - %v", err)
	}
	hasIdSlice = hasIdSlice[:0]
	for x := 1; x <= 4000; x++ {
		hasIdSlice = append(hasIdSlice, &HasId{Id: int64(x)})
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		err := g.Get(hasIdSlice[0])
		if err != nil {
			t.Errorf("Error fetching id #0 - %v", err)
		}
		wg.Done()
	}()
	go func() {
		err := g.GetMulti(hasIdSlice[1:1500])
		if err != nil {
			t.Errorf("Error fetching ids 1 through 1499 - %v", err)
		}
		wg.Done()
	}()
	go func() {
		err := g.GetMulti(hasIdSlice[1500:])
		if err != nil {
			t.Errorf("Error fetching id #1500 through 4000 - %v", err)
		}
		wg.Done()
	}()
	wg.Wait()
	for x, hi := range hasIdSlice {
		if hi.Name != "Race" {
			t.Errorf("Object #%d not fetched properly, fetched instead - %v", x, hi)
		}
	}
}

func TestPutGet(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	g := FromContext(c)

	key, err := g.Put(&PutGet{ID: 12, Value: 15})
	if err != nil {
		t.Fatal(err)
	}
	if key.IntID() != 12 {
		t.Fatal("ID should be 12 but is", key.IntID())
	}

	// Datastore Get
	dsPutGet := &PutGet{}
	err = datastore.Get(c,
		datastore.NewKey(c, "PutGet", "", 12, nil), dsPutGet)
	if err != nil {
		t.Fatal(err)
	}
	if dsPutGet.Value != 15 {
		t.Fatal("dsPutGet.Value should be 15 but is",
			dsPutGet.Value)
	}

	// Goon Get
	goonPutGet := &PutGet{ID: 12}
	err = g.Get(goonPutGet)
	if err != nil {
		t.Fatal(err)
	}
	if goonPutGet.ID != 12 {
		t.Fatal("goonPutGet.ID should be 12 but is", goonPutGet.ID)
	}
	if goonPutGet.Value != 15 {
		t.Fatal("goonPutGet.Value should be 15 but is",
			goonPutGet.Value)
	}
}

func TestMultis(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatalf("Could not start aetest - %v", err)
	}
	defer c.Close()
	n := FromContext(c)

	testAmounts := []int{1, 999, 1000, 1001, 1999, 2000, 2001, 2510}
	for _, x := range testAmounts {
		memcache.Flush(c)
		objects := make([]*HasId, x)
		for y := 0; y < x; y++ {
			objects[y] = &HasId{Id: int64(y + 1)}
		}
		if _, err := n.PutMulti(objects); err != nil {
			t.Fatalf("Error in PutMulti for %d objects - %v", x, err)
		}
		n.FlushLocalCache() // Put just put them in the local cache, get rid of it before doing the Get
		if err := n.GetMulti(objects); err != nil {
			t.Fatalf("Error in GetMulti - %v", err)
		}
	}

	// do it again, but only write numbers divisible by 100
	for _, x := range testAmounts {
		memcache.Flush(c)
		getobjects := make([]*HasId, 0, x)
		putobjects := make([]*HasId, 0, x/100+1)
		keys := make([]*datastore.Key, x)
		for y := 0; y < x; y++ {
			keys[y] = datastore.NewKey(c, "HasId", "", int64(y+1), nil)
		}
		if err := n.DeleteMulti(keys); err != nil {
			t.Fatalf("Error deleting keys - %v", err)
		}
		for y := 0; y < x; y++ {
			getobjects = append(getobjects, &HasId{Id: int64(y + 1)})
			if y%100 == 0 {
				putobjects = append(putobjects, &HasId{Id: int64(y + 1)})
			}
		}

		_, err := n.PutMulti(putobjects)
		if err != nil {
			t.Fatalf("Error in PutMulti for %d objects - %v", x, err)
		}
		n.FlushLocalCache() // Put just put them in the local cache, get rid of it before doing the Get
		err = n.GetMulti(getobjects)
		if err == nil && x != 1 { // a test size of 1 has no objects divisible by 100, so there's no cache miss to return
			t.Errorf("Should be receiving a multiError on %d objects, but got no errors", x)
			continue
		}

		merr, ok := err.(appengine.MultiError)
		if ok {
			if len(merr) != len(getobjects) {
				t.Errorf("Should have received a MultiError object of length %d but got length %d instead", len(getobjects), len(merr))
			}
			for x := range merr {
				switch { // record good conditions, fail in other conditions
				case merr[x] == nil && x%100 == 0:
				case merr[x] != nil && x%100 != 0:
				default:
					t.Errorf("Found bad condition on object[%d] and error %v", x+1, merr[x])
				}
			}
		} else if x != 1 {
			t.Errorf("Did not return a multierror on fetch but when fetching %d objects, received - %v", x, merr)
		}
	}
}
