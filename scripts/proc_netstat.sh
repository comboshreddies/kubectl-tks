#!/bin/sh
gawk 'function y(s,p,l){return strtonum("0x"substr(s,p,l))} function x(a){return y(a,7,2)"."y(a,5,2)"."y(a,3,2)"."y(a,1,2)":"y(a,10,4);}{print x($2)" "x($3)" "y($4,1,4)}'
