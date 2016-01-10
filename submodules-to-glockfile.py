#!/usr/bin/python

import re
import subprocess

def main():
    source = open(".gitmodules").read()
    paths = re.findall(r"path = (.*)", source)

    print "github.com/localhots/satan {}".format(path_sha1("."))
    for path in paths:
        print "{repo} {sha}".format(
            repo = path[7:],
            sha = path_sha1(path)
        )

def path_sha1(path):
    cmd = "cd {} && git rev-parse HEAD".format(path)
    sp = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
    sha1 = sp.stdout.read()[:-1]

    return sha1

if __name__ == "__main__": main()
