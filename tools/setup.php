<?php
# termination
function fail(string $msg): never {
	fwrite(STDERR, basename(__FILE__) . ": [ERROR] " . trim($msg) . "\n");
	exit(1);
}

# number of functions
if(!isset($argv[1]))
	define("N", 10);
elseif(is_numeric($argv[1])) {
	define("N", (int)$argv[1]);

	if(N < 3 || N > 32) {
		fail("Parameter out of range: " . N);
	}
}
else
	fail("Invalid parameter \"$argv[1]\"");
?>
