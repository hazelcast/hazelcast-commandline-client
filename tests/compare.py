#! /usr/bin/env python3

import sys


def compare_lines():
    line_count = 0
    while True:
        line = sys.stdin.readline().strip()
        if line == "":
            break
        line_count += 1
        # assumes the line format in: kN\tvN
        k, v = tuple(line.split("\t", 1))
        n = k[1:]
        expected = f"v{n}"
        if v != expected:
            raise Exception(f"line {line_count}: expected: {expected}, got: {v}")
    return line_count


print(compare_lines())
