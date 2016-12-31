Go Ledger parser
----------------

*Short term goal*: parse relatively complex Ledger files, and provide
an abstract syntax tree (a full programmatic representation of the
file), to be able to tweak some parts programmatically, and then write
back the files to disk.

* `ledgerfmt` is the first use of this library, a tool similar to
  `gofmt` that parses the file, indents and aligns according to
  conventions, and outputs the file back, without any semantic changes
  or interpretation of the data.

*Longer term goal*: do the mathematical computations of the original
Ledger program.

Layers of analysis
------------------

Several layers are presented in this library:

1. *Lexer*: where each piece of text in the file is mapped to a
   function, like `2016/09/09` to an `itemDate`, `CAD` to an
   `itemCommodity`, `20.00` to an `itemQuantity`, etc...

The current implementation interpretes:

```
2016/09/09 * Kentucky Friends Beef  ; Never go back there
  Expenses:Restaurants    20.00 CAD
  Assets:Cash             CAD -20.00    ; That hurt


2016/09/09 ! Payee
  Expenses:Misc    20.00 CAD
  Assets:Cash  ; Woah, not sure
```

and lexes it to:

```
itemEOL("\n")
itemDate("2016/09/09")
itemSpace(" ")
itemAsterisk("*")
itemSpace(" ")
itemString("Kentucky Friends Beef  ")
itemNote("; Never go back there")
itemEOL("\n")
itemSpace("  ")
itemAccountName("Expenses:Restaurants")
itemSpace("    ")
itemQuantity("20.00")
itemSpace(" ")
itemCommodity("CAD")
itemEOL("\n")
itemSpace("  ")
itemAccountName("Assets:Cash")
itemSpace("             ")
itemCommodity("CAD")
itemSpace(" ")
itemNeg("-")
itemQuantity("20.00")
itemEOL("\n")
itemEOL("\n")
itemEOL("\n")
itemDate("2016/09/09")
itemSpace(" ")
itemExclamation("!")
itemSpace(" ")
itemString("Payee")
itemEOL("\n")
itemSpace("  ")
itemAccountName("Expenses:Misc")
itemSpace("    ")
itemQuantity("20.00")
itemSpace(" ")
itemCommodity("CAD")
itemEOL("\n")
itemSpace("  ")
itemAccountName("Assets:Cash")
itemSpace("  ")
itemNote("; Woah, not sure")
itemEOL("\n")
itemEOF("")
```

2. *Parser*: node structure, or object-tree view of the Ledger file.

The parser transforms the preceding items into a structure that looks approximately like:

```
Root.ListNode.Nodes: [
- XactNode:
    XactPreSpace: "\n"
    IsCleared: true
    IsPending: false
    Description: "Kentucky Friends Beef"
    NotePreSpace: " "
    Note: "; Never go back there"
    Postings:
    - PostingNode:
        AccountPreSpace: "  "
        Account: "Expenses:Restaurants"
        AmountPreSpace: "    "
        AmountQuantity: "20.00"
        AmountNegative: false
        AmountCommodity: "CAD"
        AmountExpression: ""
        NotePreSpace: ""
        Note: ""
    - PostingNode:
        AccountPreSpace: "  "
        Account: "Assets:Cash"
        AmountPreSpace: "             "
        Amount:
          Decimal() = Decimal("-20.00")
          Quantity: "20.00"
          Negative: true
          Commodity: "CAD"
        AmountExpression: ""
        NotePreSpace: "    "
        Note: "; That hurt"
- XactNode
    XactPreSpace: "\n\n"
    IsCleared: false
    IsPending: true
    Description: "Payee"
    NotePreSpace: ""
    Note: ""
    Postings:
    - PostingNode:
        AccountPreSpace: "  "
        Account: "Expenses:Misc"
        AmountPreSpace: "    "
        Amount:
          Decimal() = Decimal("20.00")
          Quantity: "20.00"
          Negative: false
          Commodity: "CAD"
        AmountExpression: ""
        NotePreSpace: ""
        Note: ""
    - PostingNode:
        AccountPreSpace: "  "
        Account: "Assets:Cash"
        AmountPreSpace: "             "
        Amount:
          Decimal() = Decimal("-20.00")
          Quantity: "20.00"
          Negative: true
          Commodity: "CAD"
        AmountExpression: ""
        NotePreSpace: "    "
        Note: "; That hurt"

References
----------
