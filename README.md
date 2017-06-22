[![Build](https://img.shields.io/travis/xconstruct/ledger.svg?style=flat-square)](https://travis-ci.org/xconstruct/ledger)
[![API Documentation](https://img.shields.io/badge/api-GoDoc-blue.svg?style=flat-square)](https://godoc.org/github.com/xconstruct/ledger/journal)
[![BSD License](https://img.shields.io/badge/license-BSD-blue.svg?style=flat-square)](http://opensource.org/licenses/BSD)

Go Ledger
================

This is a fork of the excellent ledger parser of
(abourget)[https://github.com/abourget/ledger] that fixes a few issues and
aims to provides a high-level API to interact with
(Ledger)[http://plaintextaccounting.org] files.

The underlying library can parse relatively complex Ledger files, and provide
an abstract syntax tree (a full programmatic representation of the
file), to be able to tweak some parts programmatically, and then write
back the files to disk.

* `ledger-go` provides a few tools to interact with Ledger files, such as
  balance reports.

* `ledgerfmt`, similar to `gofmt`, parses the input file, indents and
  aligns according to conventions, and outputs the file back, without
  any semantic changes or interpretation of the data.

* `ledger2json` parses your Ledger file and outputs a `.json` file,
  which you can manipulate with any software.

Installation
============

```bash
go get -u github.com/xconstruct/ledger/cmd/...
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
