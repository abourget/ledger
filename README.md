Go Ledger parser
================

**Short term goal**: parse relatively complex Ledger files, and provide
an abstract syntax tree (a full programmatic representation of the
file), to be able to tweak some parts programmatically, and then write
back the files to disk.

* `ledgerfmt`, similar to `gofmt`, parses the input file, indents and
  aligns according to conventions, and outputs the file back, without
  any semantic changes or interpretation of the data.

* `ledger2json` parses your Ledger file and outputs a `.json` file,
  which you can manipulate with any software.

* `json2ledger` will read the same file, and produce a .ledger file,
  properly formatted (not yet implemented)


**Longer term goal**: do the mathematical computations of the original
Ledger program.


`ledger2json`
-------------

For a simple example, see the `parse_test.go` file. Here is an excerpt:

```ledger
; Top level comment

2016/09/09 = 2016-09-10 * Kentucky Friends Beef      ; Never go back there
   ; Some more notes for the transaction
  Expenses:Restaurants    20.00 CAD
  Assets:Cash             CAD -20.00    ; That hurt


2016/09/09 ! Payee
  ; Transaction notes
  Expenses:Misc    20.00 CAD
  Assets:Cash  ; Woah, not sure
 ; Here again
  ; And yet another note for this posting.

2016/09/10 Desc
  A  - $ 23
  B  23 $ @ 2 CAD
```

outputs:

```json
{
  "NodeType": 1,
  "Pos": 0,
  "Nodes": [
    {
      "NodeType": 4,
      "Pos": 0,
      "Comment": "; Top level comment"
    },
    {
      "NodeType": 5,
      "Pos": 20,
      "Space": "\n"
    },
    {
      "NodeType": 2,
      "Pos": 21,
      "Date": "2016-09-09T00:00:00Z",
      "EffectiveDate": "2016-09-10T00:00:00Z",
      "Description": "Kentucky Friends Beef      ",
      "IsPending": false,
      "IsCleared": true,
      "NotePreSpace": "",
      "Note": "; Never go back there\n; Some more notes for the transaction",
      "Postings": [
        {
          "NodeType": 3,
          "Pos": 139,
          "AccountPreSpace": "",
          "Account": "Expenses:Restaurants",
          "AccountPostSpace": "    ",
          "Amount": {
            "NodeType": 6,
            "Pos": 163,
            "Raw": "20.00 CAD",
            "Quantity": "20.00",
            "Negative": false,
            "Commodity": "CAD",
            "ValueExpr": ""
          },
          "BalanceAssertion": "",
          "BalanceAssignment": "",
          "Price": null,
          "PriceIsForWhole": false,
          "LotDate": "0001-01-01T00:00:00Z",
          "LotPrice": null,
          "NotePreSpace": "",
          "Note": ""
        },
        {
          "NodeType": 3,
          "Pos": 175,
          "AccountPreSpace": "",
          "Account": "Assets:Cash",
          "AccountPostSpace": "             ",
          "Amount": {
            "NodeType": 6,
            "Pos": 199,
            "Raw": "CAD -20.00",
            "Quantity": "20.00",
            "Negative": true,
            "Commodity": "CAD",
            "ValueExpr": ""
          },
          "BalanceAssertion": "",
          "BalanceAssignment": "",
          "Price": null,
          "PriceIsForWhole": false,
          "LotDate": "0001-01-01T00:00:00Z",
          "LotPrice": null,
          "NotePreSpace": "",
          "Note": "; That hurt"
        }
      ]
    },
    {
      "NodeType": 5,
      "Pos": 225,
      "Space": "\n\n"
    },
    {
      "NodeType": 2,
      "Pos": 227,
      "Date": "2016-09-09T00:00:00Z",
      "EffectiveDate": "0001-01-01T00:00:00Z",
      "Description": "Payee",
      "IsPending": true,
      "IsCleared": false,
      "NotePreSpace": "",
      "Note": "; Transaction notes",
      "Postings": [
        {
          "NodeType": 3,
          "Pos": 270,
          "AccountPreSpace": "",
          "Account": "Expenses:Misc",
          "AccountPostSpace": "    ",
          "Amount": {
            "NodeType": 6,
            "Pos": 287,
            "Raw": "20.00 CAD",
            "Quantity": "20.00",
            "Negative": false,
            "Commodity": "CAD",
            "ValueExpr": ""
          },
          "BalanceAssertion": "",
          "BalanceAssignment": "",
          "Price": null,
          "PriceIsForWhole": false,
          "LotDate": "0001-01-01T00:00:00Z",
          "LotPrice": null,
          "NotePreSpace": "",
          "Note": ""
        },
        {
          "NodeType": 3,
          "Pos": 299,
          "AccountPreSpace": "",
          "Account": "Assets:Cash",
          "AccountPostSpace": "  ",
          "Amount": null,
          "BalanceAssertion": "",
          "BalanceAssignment": "",
          "Price": null,
          "PriceIsForWhole": false,
          "LotDate": "0001-01-01T00:00:00Z",
          "LotPrice": null,
          "NotePreSpace": "",
          "Note": "; Woah, not sure\n; Here again\n; And yet another note for this posting."
        }
      ]
    },
    {
      "NodeType": 5,
      "Pos": 386,
      "Space": "\n"
    },
    {
      "NodeType": 2,
      "Pos": 387,
      "Date": "2016-09-10T00:00:00Z",
      "EffectiveDate": "0001-01-01T00:00:00Z",
      "Description": "Desc",
      "IsPending": false,
      "IsCleared": false,
      "NotePreSpace": "",
      "Note": "",
      "Postings": [
        {
          "NodeType": 3,
          "Pos": 405,
          "AccountPreSpace": "",
          "Account": "A",
          "AccountPostSpace": "  ",
          "Amount": {
            "NodeType": 6,
            "Pos": 408,
            "Raw": "- $ 23",
            "Quantity": "23",
            "Negative": true,
            "Commodity": "$",
            "ValueExpr": ""
          },
          "BalanceAssertion": "",
          "BalanceAssignment": "",
          "Price": null,
          "PriceIsForWhole": false,
          "LotDate": "0001-01-01T00:00:00Z",
          "LotPrice": null,
          "NotePreSpace": "",
          "Note": ""
        },
        {
          "NodeType": 3,
          "Pos": 417,
          "AccountPreSpace": "",
          "Account": "B",
          "AccountPostSpace": "  ",
          "Amount": {
            "NodeType": 6,
            "Pos": 420,
            "Raw": "23 $",
            "Quantity": "23",
            "Negative": false,
            "Commodity": "$",
            "ValueExpr": ""
          },
          "BalanceAssertion": "",
          "BalanceAssignment": "",
          "Price": {
            "NodeType": 6,
            "Pos": 426,
            "Raw": " 2 CAD",
            "Quantity": "2",
            "Negative": false,
            "Commodity": "CAD",
            "ValueExpr": ""
          },
          "PriceIsForWhole": false,
          "LotDate": "0001-01-01T00:00:00Z",
          "LotPrice": null,
          "NotePreSpace": "",
          "Note": ""
        }
      ]
    }
  ]
}
```

Shortcomings
------------

This implementation has a few limitations compared to the C++ version:

* It does not yet support all top-level constructs, like "account",
  "alias", "P", "D", "year" / "Y", etc.. Most of those should be
  simple to implement.
* It does not yet understand tags. They are only considered comments.
* It does not yet implement the `value_expr` language that allows you
  to do complex math computations directly in the postings of your
  transactions. It merely store the string text of the expression,
  PROVIDED it is enclosed in parenthesis, e.g. `(123 + 2 * 3 USD)`.
* Also note that the current implementation does not validate any
  balances. It merely acts on the text of the file.


References
----------

* The lexer is heavily based on Go's `text/template`. Also this talk by Rob Pike inspired me: https://www.youtube.com/watch?v=HxaD_trXwRE
* https://github.com/howeyc/ledger for a simpler Go parser, but with some other features.
* http://plaintextaccounting.org/ for more fun about Ledger files
