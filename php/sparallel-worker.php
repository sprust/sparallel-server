<?php

declare(strict_types=1);

while (true) {
    $data = fread(STDIN, 1024);

    if ($data === false) {
        usleep(100);

        continue;
    }

    fflush(STDOUT);
    fwrite(STDOUT, "pong: $data");
    fflush(STDIN);
}
