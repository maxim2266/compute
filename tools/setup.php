<?php
# termination
function fail(string $msg): never {
	fwrite(STDERR, basename(__FILE__) . ": [ERROR] " . trim($msg) . PHP_EOL);
	exit(1);
}

# number of functions
define("N", isset($argv[1]) ? intval($argv[1]) : 10);

if(N < 3 || N > 20)
	fail("Invalid parameter: \"{$argv[1]}\"");
?>
