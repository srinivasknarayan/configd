// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: LGPL-2.1-only

package session_test

import (
	"testing"

	. "github.com/danos/configd/session/sessiontest"
	"github.com/danos/utils/pathutil"
)

// Containers
const insidecontainer = "insidecontainer"
const remotecontainer = "remote-container"
const augmentcontainer = "augment-container"
const secondaugmentcontainer = "second-augment-container"
const testusescontainer = "testUses"

// Container Paths
var insidecontainerpath = pathutil.CopyAppend(testcontainerpath, insidecontainer)
var remotecontainerpath = []string{remotecontainer}
var augmentcontainerpath = pathutil.CopyAppend(remotecontainerpath, augmentcontainer)
var secondaugmentcontainerpath = pathutil.CopyAppend(remotecontainerpath, secondaugmentcontainer)
var testusescontainerpath = []string{testusescontainer}

// Leafs
const localimplicit = "localimplicit"
const localexplicit = "localexplicit"
const localdependent = "localdependent"
const remote = "remote"
const remotedependent = "remote-dependent"
const dependentref = "dependentref"
const insideleaf = "insideleaf"
const augmentleaf = "augment-leaf"
const otherleaf = "other-leaf"
const secondleaf = "second-leaf"
const testg1 = "testG1"
const testg2 = "testG2"
const testuses2leaf = "testUses2Leaf"
const target = "target"

// Leaf Paths
var localimplicitpath = pathutil.CopyAppend(testcontainerpath, localimplicit)
var localexplicitpath = pathutil.CopyAppend(testcontainerpath, localexplicit)
var localdependentpath = pathutil.CopyAppend(testcontainerpath, localdependent)
var remotepath = pathutil.CopyAppend(testcontainerpath, remote)
var remotedependentpath = pathutil.CopyAppend(testcontainerpath, remotedependent)
var dependentrefpath = pathutil.CopyAppend(testcontainerpath, dependentref)
var insideleafpath = pathutil.CopyAppend(insidecontainerpath, insideleaf)

var augmentleafpath = pathutil.CopyAppend(augmentcontainerpath, augmentleaf)
var otherleafpath = pathutil.CopyAppend(augmentcontainerpath, otherleaf)
var secondleafpath = pathutil.CopyAppend(secondaugmentcontainerpath, secondleaf)

var targetpath = pathutil.CopyAppend(testusescontainerpath, target)
var testuses2leafpath = pathutil.CopyAppend(testusescontainerpath, testuses2leaf)

var testusestestg1path = pathutil.CopyAppend(testusescontainerpath, testg1)
var testusestestg2path = pathutil.CopyAppend(testusescontainerpath, testg2)

// Test different feature/if-feature statements, when capabilities are all
// enabled
func TestFeatureAllCapabilitiesEnabled(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", true},
		{"", localexplicitpath, "explicit", true},
		{"", localdependentpath, "dependent", true},
		{"", remotepath, "remote", true},
		{"", remotedependentpath, "remotelydependent", true},
		{"", dependentrefpath, "localandremotedependent", true},
		{"", insideleafpath, "inside-leaf-data", true},
		{"", augmentleafpath, "AugmentLeafData", true},
		{"", secondleafpath, "SecondLeafTestData", true},
		{"", testuses2leafpath, "Testdata", true},
		{"", testusestestg1path, "G1TestData", true},
		{"", testusestestg2path, "G2TestData", true},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsAll")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature,
		DELETE_AND_COMMIT)
	srv.Cleanup()
	sess.Kill()
}

// Test that when a local feature is disabled, any schema tree nodes specifying
// it in an if-feature statement is omitted from the schema tree
func TestFeatureLocalFeatureDisabled(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", true},
		{"", localexplicitpath, "explicit", true},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsLocalFeatureDisabled")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature,
		SET)
	srv.Cleanup()
	sess.Kill()
}

// Test that a feature in a remote module, that is disabled, correctly
// omits a node from the schema tree
func TestFeatureRemoteFeatureDisabled(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", true},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", true},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsRemoteFeatureDisabled")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature,
		SET)
	srv.Cleanup()
	sess.Kill()
}

// Test that features dependent on other features are correctly disabled if
// one of the other features are disabled.
func TestFeatureDependentOnDisabledFeature(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", true},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", true},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsDependentOnDisabledFeature")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature,
		SET)
	srv.Cleanup()
	sess.Kill()
}

func TestFeatureEnabledChildDisabledByParent(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", true},
		{"", augmentleafpath, "AugmentLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsInsideContainerOff")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature,
		SET)
	srv.Cleanup()
	sess.Kill()

}

// Test that an if-feature of an augment node will disable the whole node and
// its children if the capability is disabled
func TestFeatureDisableAugment(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", true},
		{"", otherleafpath, "OtherLeafData", true},
		{"", secondleafpath, "SecondLeafTestData", true},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsDisableAugmentNode")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature, SET)
	srv.Cleanup()
	sess.Kill()
}

// Test that an if-feature of a container in an augment node will only disable
// the container and its children when.
func TestFeatureDisableAugmentContainer(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", true},
		{"", otherleafpath, "OtherLeafData", true},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsDisableAugmentContainer")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature, SET)
	srv.Cleanup()
	sess.Kill()
}

// Test that an if-feature of a leaf in an augment node will only disable
// the leaf.
func TestFeatureDisableAugmentLeaf(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", true},
		{"", otherleafpath, "OtherLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", false},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "",
		"testdata/featureValid/capsDisableAugmentLeaf")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature, SET)
	srv.Cleanup()
	sess.Kill()
}

// Test that an if-feature of a uses will disable the use of the grouping
func TestFeatureDisableUses(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", false},
		{"", otherleafpath, "OtherLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", true},
		{"", testusestestg2path, "G2TestData", true},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "", "testdata/featureValid/capsDisableUses")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature, SET)
	srv.Cleanup()
	sess.Kill()
}

// Test that an if-feature of a leaf in a grouping included via a uses
// will be disabled
func TestFeatureDisableGroupingLeafNode(t *testing.T) {
	tblFeature := []ValidateOpTbl{
		{"", localimplicitpath, "implicit", false},
		{"", localexplicitpath, "explicit", false},
		{"", localdependentpath, "dependent", false},
		{"", remotepath, "remote", false},
		{"", remotedependentpath, "remotelydependent", false},
		{"", dependentrefpath, "localandremotedependent", false},
		{"", insideleafpath, "inside-leaf-data", false},
		{"", augmentleafpath, "AugmentLeafData", false},
		{"", otherleafpath, "OtherLeafData", false},
		{"", secondleafpath, "SecondLeafTestData", false},
		{"", testuses2leafpath, "Testdata", false},
		{"", testusestestg1path, "G1TestData", false},
		{"", testusestestg2path, "G2TestData", true},
	}

	srv, sess := TstStartupSchemaDir(t, "testdata/featureValid", "", "testdata/featureValid/capsDisableG2")
	ValidateOperationTable(t, sess, srv.Ctx, tblFeature, SET)
	srv.Cleanup()
	sess.Kill()
}
