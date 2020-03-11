#!/usr/bin/python

import sys

rc = 1

if len(sys.argv) == 4:
    separator=sys.argv[2]
    token_list = sys.argv[3]
    if sys.argv[1] == "next-token":
        split_idx = token_list.find(separator)
        if split_idx != -1:
            token = token_list[:split_idx]
        else:
            token = token_list
        sys.stdout.write(token)
        rc = 0
    elif sys.argv[1] == "remove-token":
        rc = 0
        split_idx = token_list.find(separator)
        if split_idx != -1:
            token_list = token_list[split_idx + len(separator):]
        else:
            token_list = ""
        sys.stdout.write(token_list)
else:
    sys.stderr.write("Syntax: tokenizer { next-token | remove-token } <separator> <string>\n")

sys.stdout.flush()
sys.stderr.flush()

exit(rc)
