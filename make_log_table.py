#!/usr/bin/python

# Copyright 2012, Kevin Ko <kevin@faveset.com>.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# Generates an 8-bit log base 2 table.

def lt(n):
  print (",".join([str(n) for i in xrange(16)]) + ",")

# Print out an inverse 8-bit log table
print "-1, 0, 1, 1,"  # log(i) for i in 0, 1, 2, 3
print "2, 2, 2, 2,"   # ... 4, 5, 6, 7
print "3, 3, 3, 3, 3, 3, 3, 3,"  # ... 8-15
lt(4)  # 16-31

# 32-63
[lt(5) for i in xrange(2)]

# 64-127
[lt(6) for i in xrange(4)]

# 128-255
[lt(7) for i in xrange(8)]
