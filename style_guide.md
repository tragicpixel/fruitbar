Style Guide
===========
This style guide covers things that gofmt doesn't automatically change.

Import Order
------------
- Level 1: First-party go standard library things
- Level 2: Imports from this project
- Level 3: Third party imports

Naming Constants
----------------
Don't use all caps and underscores.
Use camel case, first letter uppercase.

Conflicting Package Names
-------------------------
If a package for this project uses the same name as a go standard library package, keep it, but change the name when importing it.