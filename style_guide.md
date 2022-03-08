Style Guide
===========
This style guide covers things that gofmt doesn't automatically change.
It's useful as a reference in case one of the use cases below comes along while I am coding.

Most importantly
----------------
Follow the local style, whatever it is. Be a team player; you will not always be the sole author of a project and different teams may use different styles.

Import Order
------------
- Level 1: First-party go standard library things
- Level 2: Imports from this project
- Level 3: Third party imports

Each level separated by an empty line.

Within each level, separate into groups that are related.
Put more important stuff first, and less important things (like helper functions, utilities, etc.) later.

Naming
------
- Keep local variable names short. Common variables like an index or loop letter may be a single letter.
- Acronyms should be in all capitals.
- Short variable names work better when the distance between their declaration and last use is short.
- Long variable names need to justify themselves: the longer they are, they more value they need to provide.
- Don't include the name of your type in the name of your variable.
- Prefer single letter variables for loops and branches.
- Prefer single words for parameters and return values.
- Prefer multiple words for functions and package level declarations.
- Prefer single words for methods(belong to a type), interfaces, and packages.

Does the extra verbosity add anything to the comprehension of the program??
- Use standard abbrevations where appropriate for a domain. (reduce cognitive load)

- Don't name your variables for their types!
    - Go is statically typed! So putting the type in the name is redundant; Go won't let you use it inappropriately anyway.
    - This doesn't improve the clarity of the code, it's just extra boilerplate to type.
    - This goes for function parameters too: don't name argument of type *Config "config" -- we already know that!! (just name it "c" or abbreviate it to "conf")
        - Could rename it to "cfg" but "conf" is a more widely used abbreviation.
        - Don't be afraid to look these things up!

- Don't use numbers in variable names usually.
    - Naming it `conf1` and `conf2` doesn't give you good information about it, try `original` and `updated` instead so the variables are less likely to be confused.

- Don't let package names steal good variable names. (you can rename them)

- Use a consistent naming style!
    - Don't have `d *sql.DB, dbase *sql.DB, DB *sql.DB and database *sql.DB` instead pick one.
    - Use the same reciever name on very method on that type. `func (p person) print() {}` : use `p` everywhere! The reader can more easily internalize the use of the reciever across the methods in this type.

Naming Constants
----------------
- Use MixedCase for constants.
- Remember that if the letter is lowercase, the constant will be private to the package.
- Constants should describe the value they hold, not HOW that value is used.

Conflicting Package Names
-------------------------
If a package for this project uses the same name as a go standard library package, keep it, but change the name when importing it.

Handle errors first/Avoid nesting
---------------------------------
- Check for an error at the beginning of the program, this simplifies the code and reduces the load on the reader.
- Avoid nesting: when an error is encountered, fail immediately and exit, instead of using else statements.
- Easy off-the-cuff test: If a line is longer than 120 columns, you probably have too much indentation going on, and it probably could be simplified.

Prefer switch to if/else
------------------------
- If you are testing multiple cases, writing it as a switch statement will result in less code.
- Remember that cases in go can be expressions.

Documentation
-------------
- Create a doc.go file if there is more than one file in a package.

Comments
--------
Comments do one of three things:
1. Explain **what** the thing does. (ideal for public symbols)
2. Explain **how** the thing does it. (commentary inside a method)
3. Explain **why** the thing is what it is. (provide context where something otherwise doesn't make sense)

- Comments on variables and constants should describe their contents, not their purpose.
    - i.e. why is the value six? Not where this constant will be used.
- For variables without an initial value, the comment should describe who is responsible for initializing this variable.
- Sometimes you'll find a better name for a variable hiding in a comment:
`var registry = make(map[string]*sql.Driver) // registry of SQL drivers` to `var sqlDrivers = ...` (comment removed)

- Any public function that is not both obvious and short must be commented.
- Any function in a library must be commented regardless of length or complexity.
- You don't need to document methods that implement an interface.
    - A comment that says "implements X interface" is essentially just telling you to go look somewhere else for the documentation.

- Use TODO(username) to annotate problems to fix later -- but you should additionally raise an issue to refactor it later, don't just ignore it.
    -Putting your username isn't a commitment to you fixing the problem, but that you might be the best person to talk to first when it comes time to fix it.

- If you find yourself commenting a piece of code because it is unrelated to the rest of the function, consider moving it into a function of its own.

Function Arguments
------------------
In instances where you are passing large numbers of arguments to a function, create a small helper type instead, and pass that to the function.
- Makes things more readable.
- Makes things more maintainable.

Don't mix `nil` and non-`nil`-able parameters in the same function signature:
- Your API should not be requiring the caller to provide parameters they don't care about.
- Design your APIs for the **default** use case!!
- Give serious consideration to how much time helper functions will save the programmer: **clear is better than concise**!
    - Use Public wrappers to hide the messy details.

### Prefer var args to slice parameters
- You can pass an empty slice, or `nil` as a slice, and the compiler will be happy -- but this means more edge cases for you to test.
- If you are only passing one value, you are forced to box it into a slice.
- Note that you can also pass `...` as the argument for var args, and this will compile.
    - Here's a tip: Make the first parameter explicit so you can't make these empty calls, and then the rest are var args. `func (first int, rest ...int)`

Quick note about error handling to remember
-------------------------------------------
The contract for error handling in Go says that you cannot make assumptions about the contents of other return values in the presence of an error.