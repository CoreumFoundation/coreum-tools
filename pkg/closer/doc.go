// Package closer ensures a clean exit for your Go app.
//
// The aim of this package is to provide an universal way to catch the event of application’s exit
// and perform some actions before it’s too late. Closer doesn’t care about the way application
// tries to exit, i.e. was that a panic or just a signal from the OS, it calls the provided methods
// for cleanup and that’s the whole point.
//
// Exit codes
//
// The exit code (for `os.Exit`) will be determined accordingly:
//
//   Event         | Default exit code
//   ------------- | -------------
//   error = nil   | 0 (success)
//   error != nil  | 1 (failure)
//   panic         | 1 (failure)
//
package closer
