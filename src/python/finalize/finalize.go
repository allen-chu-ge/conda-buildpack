package finalize

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
}

type Manifest interface {
	RootDir() string
}

type Stager interface {
	BuildDir() string
	DepDir() string
	LinkDirectoryInDepDir(string, string) error
}

type Finalizer struct {
	Stager   Stager
	Command  Command
	Log      *libbuildpack.Logger
	Manifest Manifest
}

func Run(f *Finalizer) error {
	// TODO: Conda

	// source $BIN_DIR/steps/mercurial

	// TODO: Uninstall removed dependencies with Pip.
	// source $BIN_DIR/steps/pip-uninstall

	if err := f.PipInstall(); err != nil {
		f.Log.Error("Unable to perform pip install: %s", err.Error())
		return err
	}

	// TODO: Make sure all Pip installed dependencies use #!/usr/bin/env python
	// $BIN_DIR/steps/rewrite-shebang

	// TODO: Support for NLTK corpora.
	// $BIN_DIR/steps/nltk

	// TODO: Django collectstatic support.
	// $BIN_DIR/steps/collectstatic

	// TODO App based hooks

	// TODO: rewrite build dir in egg links to runtime $DEPS_DIR/app so things can be found

	// TODO: cache everything

	return nil
}

func (f *Finalizer) PipInstall() error {
	// if [ -f $CACHE_DIR/egg-path-prefix ]; then
	//   prefix=$(cat $CACHE_DIR/egg-path-prefix)

	//   set +e
	//   # delete any existing egg links, to uninstall existing installations.
	//   find $DEPS_DIR/$DEPS_IDX/python/lib/python*/site-packages/ -name "*.egg-link" -delete 2> /dev/null
	//   find $DEPS_DIR/$DEPS_IDX/python/lib/python*/site-packages/ -name "*.pth" -print0 2> /dev/null | xargs -r -0 -n 1 sed -i -e "s#$prefix#$DEPS_DIR/$DEPS_IDX#" &> /dev/null

	//   # Support for the above, for PyPy.
	//   find $DEPS_DIR/$DEPS_IDX/python/lib-python/*/site-packages/ -name "*.egg-link" -print0 2> /dev/null | xargs -r -0 -n 1 sed -i -e "s#$prefix#$DEPS_DIR/$DEPS_IDX#" &> /dev/null
	//   find $DEPS_DIR/$DEPS_IDX/python/lib-python/*/site-packages/ -name "*.pth" -print0 2> /dev/null | xargs -r -0 -n 1 sed -i -e "s#$prefix#$DEPS_DIR/$DEPS_IDX#" &> /dev/null
	//   set -e
	// fi

	exists, err := libbuildpack.FileExists(filepath.Join(f.Stager.BuildDir(), "vendor"))
	if err != nil {
		return err
	}
	// TODO OUTPUT for all below should '| cleanup | indent'
	if exists {
		if err := f.Command.Execute(f.Stager.BuildDir(), os.Stdout, ioutil.Discard, "pip", "install", "-r", "requirements.txt", "--exists-action=w", "--src="+filepath.Join(f.Stager.DepDir(), "src"), "--no-index", "--find-links=file:/"+f.Stager.BuildDir()+"/vendor"); err != nil {
			return err
		}
		if err := libbuildpack.CopyFile(filepath.Join(f.Stager.BuildDir(), "requirements.txt"), filepath.Join(f.Stager.DepDir(), "python", "requirements-declared.txt")); err != nil {
			return err
		}
		f2, err := os.Create(filepath.Join(f.Stager.DepDir(), "python", "requirements-installed.txt"))
		if err != nil {
			return err
		}
		if err := f.Command.Execute(f.Stager.BuildDir(), f2, ioutil.Discard, "pip", "freeze", "--find-links=file:/"+f.Stager.BuildDir()+"/vendor", "--disable-pip-version-check"); err != nil {
			return err
		}
		f2.Close()
	} else {
		if err := f.Command.Execute(f.Stager.BuildDir(), os.Stdout, ioutil.Discard, "pip", "install", "-r", "requirements.txt", "--exists-action=w", "--src="+filepath.Join(f.Stager.DepDir(), "src")); err != nil {
			return err
		}
		if err := libbuildpack.CopyFile(filepath.Join(f.Stager.BuildDir(), "requirements.txt"), filepath.Join(f.Stager.DepDir(), "python", "requirements-declared.txt")); err != nil {
			return err
		}
		f2, err := os.Create(filepath.Join(f.Stager.DepDir(), "python", "requirements-installed.txt"))
		if err != nil {
			return err
		}
		if err := f.Command.Execute(f.Stager.BuildDir(), f2, ioutil.Discard, "pip", "freeze"); err != nil {
			return err
		}
		f2.Close()
	}

	if err := f.Stager.LinkDirectoryInDepDir(filepath.Join(f.Stager.DepDir(), "python", "bin"), "bin"); err != nil {
		return err
	}

	// # Record for future reference.
	// echo $DEPS_DIR/$DEPS_IDX > "$CACHE_DIR/egg-path-prefix"

	return nil
}
