# Contributing

1. Clone this repository

    ```
    go get -d github.com/howtowhale/dvm
    cd $GOPATH/src/github.com/howtowhale/dvm
    ```

1. Add your fork as a remote:

    ```
    git remote add fork git@github.com:USERNAME/dvm.git
    ```

1. Build `dvm`:

    ```
    make
    ```

1. Run tests:

    ```
    make test
    ```

1. Try out your local build:

    ```
    source ./dvm.sh
    dvm --version
    ```
