Lyft Runtime Configuration
=======

This repo contains data that is distributed to all instances, quickly. 

The contents of the /data/ directory in this repository are mirrored on
all EC2 instances, in the filesystem as /srv/runtime_data/current/.   There is no
release process: commits to master are distributed to all instances
immediately (within seconds).

Note that runtime contents are distributed to *all* EC2 instances
equally:  so, if you want runtime content to be different in production,
staging, onebox, etc., you must create environment-specific content
and manage defaults, overrides, etc., yourself.   See instant-server
for examples on how to do this.

Intended Use
=======
The runtime system is meant to support small amounts of data, such
as feature flags, kill switches, regional configuration, experiment
settings, etc.  Individual files should typically contain a single key/value pair
(filename as key, content as value), in order to allow for GUI tools to
operate against this repository with minimal git conflicts.

*There is a maximum file size limit of 1 megabyte.*

Data Validation
=======
Before moving runtime content to all instances, "make test" will be invoked.
If that fails, no data will be moved.   This process is also run by jenkins
in order to notify people that the runtime "build" is failing.

Runtime in Devbox
=======
To use runtime in your local devbox (your laptop), do the following:

    1. git clone git@github.com:lyft/goruntime.git
    2. cd devbox 
    3. ./service start runtime

You will see /srv/runtime_data/current/ update every 10 seconds in all your devbox containers,
based on the content of your runtime.git checkout.  You can watch this all happening
by looking at /var/log/supervisor/local_runtime_sync.out.log in your runtime container,
and by looking at /var/log/supervisor/runtime.out.log on every other container.

Best Practices
=======

Code that uses runtime configuration should follow these best practices:

1. Do not assume that /srv/runtime_data/current will always be present.  
   Your code should work even if /srv/runtime_data/current doesn't exist.

1. Sane defaults.   Your code should do the "right" thing if /srv/runtime_data/current/
   isn't present.  If by default a feature should be on, then its runtime
default should be on.  If by default it should be off, then its runtime default
should be off.  This means that the defaults should change over time.

1. Your code should validate the content of runtime data.  Even where we have
   robust type-checking earlier in the runtime pipeline (in the UI, as part of
the "make test" check, etc), the code should always check that the data matches
expectations.

1. Use the standard libraries for accessing runtime.  For example, https://github.com/lyft/python-lyft-stdlib/blob/master/lyft/goruntime.py

