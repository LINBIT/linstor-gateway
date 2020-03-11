#!/usr/bin/python

import os
import sys

KEY_IP = "OCF_RESKEY_ip"
KEY_PORT = "OCF_RESKEY_portno"

def main():
    match_address = os.environ.get(KEY_IP)
    match_port = os.environ.get(KEY_PORT)

    if len(sys.argv) == 2:
        if match_address is not None and match_port is not None:
            if sys.argv[1] == "save":
                process_send(match_address, match_port)
            elif sys.argv[1] == "query":
                process_query(match_address, match_port)
        else:
            sys.stderr.write("Error: Required environment variables are not defined:\n")
            if match_address is None:
                sys.stderr.write("    " + KEY_IP + "\n")
            if match_port is None:
                sys.stderr.write("    " + KEY_PORT + "\n")
    else:
        sys.stderr.write("Arguments: save | query\n")

def process_send(match_address, match_port):
    dst_string = ""
    for line in sys.stdin:
        fields = line.split()
        if len(fields) >= 6:
            src_address, src_port = split_address_and_port(fields[3])
            if src_address == match_address and src_port == match_port:
                if len(dst_string) >= 1:
                    dst_string += ","
                dst_string += fields[4]
        else:
            sys.stderr.write("WARNING: Unparsable input line:\n    ")
            sys.stderr.write(line)

    sys.stdout.write(dst_string + "\n")

def process_query(match_address, match_port):
    process = True
    for line in sys.stdin:
        # Process the first line
        if process:
            value_idx = line.find("value=")
            if value_idx != -1:
                value = line[value_idx + 6:]
                dst_string = value.split(',')
                for dst_address_and_port in dst_string:
                    sys.stdout.write(match_address + ":" + match_port + " " + dst_address_and_port + "\n")
                # Read and discard the rest of the lines
                process = False

def split_address_and_port(address_and_port):
    address = None
    port = None
    split_idx = address_and_port.rfind(':')
    if split_idx != -1:
        address = address_and_port[:split_idx]
        port = address_and_port[split_idx + 1:]
    return address, port

if __name__ == "__main__":
    main()
