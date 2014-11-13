glululemon
==========

Glue packages for tying blockchains and filesystems (modules) to decervers.

This repository is tightly coupled to (`decerver-interfaces`)[https://github.com/eris-ltd/decerver-interfaces)

Note, for a given module to be used with the decerver, it must satisfy the Module interface in `decerver-interfaces/modules` and wrap an object which satisfies the particular interface for the functionality the module provides. For example, `IpfsModule` satisfies both `Module` and `FileSystem`, and wraps an object (`Ipfs`) which satisfies only `FileSystem`. This is necessary so we can bind `Ipfs` to the javascript virtual machine without also exposing the configuration and booting functions in the `IpfsModule` object, but also use the `IpfsModule` as a standalone wrapper for using Ipfs alone or in other programs.
