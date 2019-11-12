// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: LGPL-2.1-only

package session_test

import (
	"testing"

	"github.com/danos/config/testutils"
	. "github.com/danos/configd/session/sessiontest"
	"github.com/danos/utils/pathutil"
)

// Declarations required in addition to those in session_test.go
const subcontainer = "subcontainer"
const level3container = "level3container"
const level4container = "level4container"
const subtrue = "subtrue"
const subfalse = "subfalse"
const level3leaf = "level3leaf"
const level4leaf = "level4leaf"

var subcontainerpath = pathutil.CopyAppend(testcontainerpath, subcontainer)
var level3containerpath = pathutil.CopyAppend(subcontainerpath, level3container)
var level4containerpath = pathutil.CopyAppend(level3containerpath, level4container)
var subfalsepath = pathutil.CopyAppend(subcontainerpath, subfalse)
var subtruepath = pathutil.CopyAppend(subcontainerpath, subtrue)
var level3leafpath = pathutil.CopyAppend(level3containerpath, level3leaf)
var level4leafpath = pathutil.CopyAppend(level4containerpath, level4leaf)

// Check that "mandatory false" statement is correctly compiled as
// "mandatory false"
func TestMandatoryFalseIsFalse(t *testing.T) {
	const schemaMandFalse = `
container testcontainer {
	leaf testleaf {
		type string;
		mandatory false;
	}
	leaf teststring {
		type string;
		mandatory false;
	}
	leaf testboolean {
		type boolean;
		mandatory false;
	}
	list testlist {
		key nodetag;
		leaf nodetag {
			type string;
		}
		leaf testleaf {
			type string;
			mandatory false;
		}
	}
	leaf-list testleaflistuser {
		type string;
	}
}
`

	tblSetMandFalse := []ValidateOpTbl{
		{"Verify set of non-mandatory", testleafpath, "foo", true},
		{"", testbooleanpath, "true", true},
		{"", testleaflistuserpath, "foo", true},
		{"", teststringpath, "foo", true},
	}
	tblDeleteMandFalse := []ValidateOpTbl{
		{"Verify delete of non-mandatory", testbooleanpath, "true", true},
		{"", testleafpath, "foo", true},
		{"", teststringpath, "foo", true},
		{"", testleaflistuserpath, "foo", true},
	}

	srv, sess := TstStartup(t, schemaMandFalse, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSetMandFalse,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandFalse,
		DELETE_AND_COMMIT)
	sess.Kill()

}

// Check that a mandatory node with only non-presence container
// ancestors, forces the container to exists with all
// mandatory nodes satisfied
func TestMandatoryInTopNonPresenceContainer(t *testing.T) {
	const schemaMandTrue = `
container testcontainer {
	leaf testleaf {
		type string;
		mandatory true;
	}
	leaf teststring {
		type string;
		mandatory true;
	}
	leaf testboolean {
		type boolean;
		mandatory false;
	}
	list testlist {
		key nodetag;
		leaf nodetag {
			type string;
		}
		min-elements 1;
	}
	leaf-list testleaflistuser {
		type string;
		min-elements 1;
	}
	list listwithmand {
		key nodetag;
		leaf nodetag {
			type string;
		}
		leaf mandlistnode {
			description "The scope of this mandatory
			leaf is limited to the parent list \"listwithmand\"";
			type string;
			mandatory true;
		}
	}
}
`
	// Ensure that initial commit will fail unless all mandatory
	// nodes are present in the config.
	mandatoryNodesConfig := testutils.Root(
		testutils.Cont("testcontainer",
			testutils.Leaf("testleaf", "foo"),
			testutils.Leaf("testboolean", "true"),
			testutils.Leaf("teststring", "foo"),
			testutils.LeafList("testleaflistuser",
				testutils.LeafListEntry("foo")),
			testutils.List("testlist",
				testutils.ListEntry("bar"))))

	// Check that a non-mandatory node can be deleted if all mandatory
	// leafs are still present.
	// Delete will fail in the absence of any mandatory leafs.
	// Being a top level, non-presence container (parent is root),
	// if any mandatory nodes are missing, commit will fail.
	tblDeleteMandTrue := []ValidateOpTbl{
		{"Verify delete of non-mandatory node with mandatory siblings",
			testbooleanpath, "true", true},
		{"Delete of mandatory node is rejected", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "bar", false},
	}

	srv, sess := TstStartup(t, schemaMandTrue, mandatoryNodesConfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandTrue,
		DELETE_AND_COMMIT)
	sess.Kill()
}

// Check that a mandatory node with a presence container
// as an ancestor, must exist whenever the presence container
// exists.
func TestMandatoryInPresenceContainer(t *testing.T) {
	const schemaMandTrue = `
container testcontainer {
	presence "To limit scope of the mandatory";
	leaf testleaf {
		type string;
		mandatory true;
	}
	leaf teststring {
		type string;
		mandatory true;
	}
	leaf testboolean {
		type boolean;
		mandatory false;
	}
	list testlist {
		key nodetag;
		leaf nodetag {
			type string;
		}
		min-elements 1;
	}
	leaf-list testleaflistuser {
		type string;
		min-elements 1;
	}
}
`

	// Ensure that commit will fail unless all mandatory
	// nodes are present in the config
	tblSetMandTrue := []ValidateOpTbl{
		{"Validate commit fails with missing mandatory nodes", testbooleanpath, "true", false},
		{"", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"All mandatory nodes now satisfied", testlistpath, "bar", true},
	}

	// Check that a non-mandatory node can be deleted if all mandatory
	// leafs are still present.
	// Delete will fail in the absence of any
	// mandatory leafs until ALL nodes AND the presence container
	// are deleted.
	tblDeleteMandTrue := []ValidateOpTbl{
		{"Verify delete of non-mandatrory node with mandatory children", testbooleanpath, "true", true},
		{"", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "bar", false},
		{"", testcontainerpath, "", true},
	}

	srv, sess := TstStartup(t, schemaMandTrue, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSetMandTrue,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandTrue,
		DELETE_AND_COMMIT)
	sess.Kill()
}

// Check that mandatory nodes in a child, non-presence container,
// have to be present before commit of anything in it's parent
// presence container.
func TestMandatoryInNonPresenceContainer(t *testing.T) {
	const schemaMandTrue = `
container testcontainer {
	presence "To limit scope of the mandatory";
	leaf testleaf {
		type string;
		mandatory true;
	}
	leaf teststring {
		type string;
		mandatory true;
	}
	leaf testboolean {
		type boolean;
		mandatory false;
	}
	list testlist {
		key nodetag;
		leaf nodetag {
			type string;
		}
		min-elements 1;
	}
	leaf-list testleaflistuser {
		type string;
		min-elements 1;
	}
	container subcontainer {
		leaf subtrue {
			type string;
			mandatory true;
		}
		leaf subfalse {
			type string;
			mandatory false;
		}
	}
}
`

	// Ensure that commit will fail unless all mandatory
	// nodes are present in the config
	tblSetMandTrue := []ValidateOpTbl{
		{"Validate commit fails with missing mandatory nodes", testbooleanpath, "true", false},
		{"", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "bar", false},
		{"All mandatory constraints satisfied", subtruepath, "true", true},
		{"", subfalsepath, "false", true},
	}

	// Check that a non-mandatory node can be deleted if all mandatory
	// leafs are still present.
	// Delete will fail in the absence of any mandatory leafs until
	// the parent non-presence container is deleted
	tblDeleteMandTrue := []ValidateOpTbl{
		{"Verify delete of non-mandatrory node with mandatory children", testbooleanpath, "true", true},
		{"", subfalsepath, "false", true},
		{"Fails the mandatory constraint", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "bar", false},
		{"", subtruepath, "true", false},
		{"Verify that presence container delete is allowed", testcontainerpath, "", true},
	}

	srv, sess := TstStartup(t, schemaMandTrue, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSetMandTrue,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandTrue,
		DELETE_AND_COMMIT)
	sess.Kill()
}

// Check that mandatory nodes in a child, presence container,
// do not have to be present before commit of anything in it's parent
// presence container.
func TestMandatoryInPresenceSubContainer(t *testing.T) {
	const schemaMandTrue = `
container testcontainer {
	presence "To limit scope of the mandatory";
	leaf testleaf {
		type string;
		mandatory true;
	}
	leaf teststring {
		type string;
		mandatory true;
	}
	leaf testboolean {
		type boolean;
		mandatory false;
	}
	list testlist {
		key nodetag;
		leaf nodetag {
			type string;
		}
		min-elements 1;
	}
	leaf-list testleaflistuser {
		type string;
		min-elements 1;
	}
	container subcontainer {
		presence "To limit scope of mandatory";
		leaf subtrue {
			type string;
			mandatory true;
		}
		leaf subfalse {
			type string;
			mandatory false;
		}
	}
}
`

	tblSetMandTrue := []ValidateOpTbl{
		{"Validate commit fails with missing mandatory nodes", testbooleanpath, "true", false},
		{"", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "bar", true},
		{"", subfalsepath, "false", false},
		{"All mandatory constraints satisfied", subtruepath, "true", true},
	}

	// Check that a non-mandatory node can be deleted if all mandatory
	// leafs are still present.
	// Delete will fail in the absence of any
	// mandatory leafs until ALL container nodes are deleted
	tblDeleteMandTrue := []ValidateOpTbl{
		{"Verify delete of non-mandatrory node with mandatory children", testbooleanpath, "true", true},
		{"", subtruepath, "true", false},
		{"", subfalsepath, "false", false},
		{"Verify subcontainer can be deleted", subcontainerpath, "", true},
		{"Fails the mandatory constraint", testleafpath, "foo", false},
		{"", teststringpath, "foo", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "bar", false},
		{"Verify that presence container delete is allowed", testcontainerpath, "", true},
	}

	srv, sess := TstStartup(t, schemaMandTrue, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSetMandTrue,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandTrue,
		DELETE_AND_COMMIT)
	sess.Kill()
}

// Check that a mandatory node multiple levels down in a
// hierarchy of non-presence containers, prevents commit
// unless the mandatory node is present
func TestMandatoryInQuadNestedContainer(t *testing.T) {
	const schemaMandTrue = `
container testcontainer {
	presence "To limit scope of the mandatory";
	leaf testleaf {
		type string;
	}
	container subcontainer {
		leaf subfalse {
			type string;
		}
		container level3container {
			leaf level3leaf {
				type string;
			}
			container level4container {
				leaf level4leaf {
					type string;
					mandatory true;
				}
			}
		}
	}
}
`

	// Ensure that commit will fail unless the mandatory
	// node, four levels down, is present
	tblSetMandTrue := []ValidateOpTbl{
		{"Reject commit due to mandatory contraint", testleafpath, "foo", false},
		{"Reject commit due to mandatory constraint", subfalsepath, "false", false},
		{"Reject commit due to mandatory constraint", level3leafpath, "level3", false},
		{"Accept commit, mandatory constraints now ssatisfied", level4leafpath, "level4", true},
	}

	// Check that non-mandatory nodes in the hierarchy can be deleted
	// so long as the mandatory
	tblDeleteMandTrue := []ValidateOpTbl{
		{"Commit allowed, mandatory still satisfied", subfalsepath, "false", true},
		{"", level3leafpath, "level3", true},
		{"", testleafpath, "foo", true},
		{"Reject commit, mandatory constraint not satisfied", level4leafpath, "level4", false},
		{"Commit allowed, presence container has been deleted", testcontainerpath, "", true},
	}

	// Do above tests in different order
	tblSetMandTrueReversed := []ValidateOpTbl{
		{"Commit success, mandatory constraint satisfied", level4leafpath, "level4", true},
		{"", level3leafpath, "level3", true},
		{"", subfalsepath, "false", true},
		{"", testleafpath, "foo", true},
	}

	tblDeleteMandTrueReversed := []ValidateOpTbl{
		{"Reject Commit, mandatory contraint enforced", level4leafpath, "level4", false},
		{"", testleafpath, "foo", false},
		{"", level3leafpath, "level3", false},
		{"", subfalsepath, "false", false},
		{"Commit success, parent presence container deleted", testcontainerpath, "", true},
	}

	srv, sess := TstStartup(t, schemaMandTrue, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSetMandTrue,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandTrue,
		DELETE_AND_COMMIT)
	sess.Kill()

	srv, sess = TstStartup(t, schemaMandTrue, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSetMandTrueReversed,
		SET_AND_COMMIT)
	ValidateOperationTable(t, sess, srv.Ctx, tblDeleteMandTrueReversed,
		DELETE_AND_COMMIT)
	sess.Kill()
}

/*
 * Validate that min-elements and max-elements constraint is correctly
 * enforced for list and leaf-list yang elements
 */
func TestMinMaxElements(t *testing.T) {

	var testlistunbounded = "testlistunbounded"
	var testlistunboundedpath = pathutil.CopyAppend(testcontainerpath, testlistunbounded)
	const schema = `
container testcontainer {
	presence "A presence container";
	list testlist {
		key nodetag;
		leaf nodetag {
			type string;
		}
		min-elements 2;
		max-elements 3;
	}
	leaf-list testleaflistuser {
		type string;
		ordered-by user;
		min-elements 1;
	}
	list testlistunbounded {
		key nodetag;
		leaf nodetag {
			type string;
		}
		min-elements 0;
		max-elements unbounded;
	}
	leaf testboolean {
		type boolean;
	}
}
`

	const configDelete = `
testcontainer {
	testboolean true
	testlist foo
	testlist bar
	testlist foobar
	testleaflistuser foo
	testleaflistuser bar
	testleaflistuser foobar
	testleaflistuser baz
	testlistunbounded foo
	testlistunbounded bar
	testlistunbounded foobar
}
`

	tblSet := []ValidateOpTbl{
		{"", testbooleanpath, "true", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testlistpath, "foo", false},
		{"", testlistpath, "bar", true}, // min-elements constraints now satisfied
		{"", testleaflistuserpath, "bar", true},
		{"", testleaflistuserpath, "foobar", true},
		{"", testleaflistuserpath, "baz", true},
		{"", testlistpath, "foobar", true},
		{"", testlistunboundedpath, "foo", true},
		{"", testlistunboundedpath, "bar", true},
		{"", testlistunboundedpath, "foobar", true},
		{"", testlistpath, "baz", false}, // max-elements exceeded; fail
	}

	tblDelete := []ValidateOpTbl{
		{"", testlistpath, "foo", true},
		{"", testleaflistuserpath, "bar", true},
		{"", testleaflistuserpath, "baz", true},
		{"", testlistunboundedpath, "foo", true},
		{"", testlistunboundedpath, "bar", true},
		{"", testlistunboundedpath, "foobar", true},
		{"", testlistpath, "bar", false}, // min-elements constraint now prevents commit
		{"", testlistpath, "foobar", false},
		{"", testleaflistuserpath, "foo", false},
		{"", testleaflistuserpath, "foobar", false},
		{"", testbooleanpath, "true", false},
		{"", testcontainerpath, "", true}, // everything now gone, commit succeeds
	}

	srv, sess := TstStartup(t, schema, emptyconfig)
	ValidateOperationTable(t, sess, srv.Ctx, tblSet, SET_AND_COMMIT)
	sess.Kill()
	srv, sess = TstStartup(t, schema, configDelete)
	ValidateOperationTable(t, sess, srv.Ctx, tblDelete, DELETE_AND_COMMIT)
	sess.Kill()
}

// TODO: (pac) Test cases for choice and anyxml once implemented.
