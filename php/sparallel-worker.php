<?php

declare(strict_types=1);

sleep(mt_rand(1, 5));

exit(0);

while (true) {
    $data = fread(STDIN, 1024);

    if ($data === false) {
        continue;
    }

    fwrite(STDOUT, "pong: $data");

    break;
}
