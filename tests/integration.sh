#!/bin/bash
# This integration test runs a collection of fixtures against the compiled
# streammon binary within the repository. The intention is to catch any errors
# within integrations and prevent regressions in functionality.

# The script should be run from the tests subdirectory within the streammon
# git repository.
BINDIR=$1

function cleanup() {
	rm test*.log &> /dev/null
}

trap cleanup EXIT SIGINT SIGKILL

# Test that the quoted string functionality works as expected.
test1() {
	EXPECTED=307
	timeout 2 \
		$BINDIR/streammon\
		-r '09:[0-9]*' \
		-f $BINDIR/tests/access.log \
		-c $BINDIR/tests/push_to_file.sh \
		-a "'#{1} #{4} #{5} #{6} #{7} #{8} #{9} #{10}' $FUNCNAME"

	assert_lc "$FUNCNAME" "$EXPECTED"
}

# Test that the quoted string functionality works as expected.
test2() {
	EXPECTED=1042
	timeout 2 \
		$BINDIR/streammon\
		-r 'Mozilla.*' \
		-f $BINDIR/tests/access.log \
		-c $BINDIR/tests/push_to_file.sh \
		-a "#{1} $FUNCNAME"

	assert_lc "$FUNCNAME" "$EXPECTED"
}

# assert_lc
# $1 == test name
# $2 == expected count in file
function assert_lc() {
	if [[ $(wc -l $BINDIR/tests/$1.log | cut -f1 -d' ') != $2 ]]; then
		echo "ERROR IN TEST $1"
		exit 1
	fi
	echo "$1 succeeded"
}

cleanup
test1
test2

