class HMREvent extends Event {
  constructor(type, file, oldFile) {
    super(type);
    this.file = file;
    this.oldFile = oldFile;
  }
}

(function () {
  class HMR extends EventTarget {
    static CHANGE = "change";
    static CREATE = "create";
    static DELETE = "delete";
    static RENAME = "rename";

    CHANGE = HMR.CHANGE;
    CREATE = HMR.CREATE;
    DELETE = HMR.DELETE;
    RENAME = HMR.RENAME;

    onCurrentPageChange(callback, options) {
      const currentFile = document.querySelector("meta[name='_serve:fname']");
      return this.onChange((ev) => {
        if (currentFile) {
          if (currentFile.content === ev.file) {
            callback(ev);
          }
        }
      }, options);
    }

    onChange(callback, options) {
      this.addEventListener(HMR.CHANGE, callback, options);
      return () => {
        this.removeEventListener(HMR.CHANGE, callback);
      };
    }

    onCreate(callback, options) {
      this.addEventListener(HMR.CREATE, callback, options);
      return () => {
        this.removeEventListener(HMR.CREATE, callback);
      };
    }

    onDelete(callback, options) {
      this.addEventListener(HMR.DELETE, callback, options);
      return () => {
        this.removeEventListener(HMR.DELETE, callback);
      };
    }

    onRename(callback, options) {
      this.addEventListener(HMR.RENAME, callback, options);
      return () => {
        this.removeEventListener(HMR.RENAME, callback);
      };
    }

    emitChanged(file) {
      this.dispatchEvent(new HMREvent(HMR.CHANGE, file));
    }

    emitCreated(file) {
      this.dispatchEvent(new HMREvent(HMR.CREATE, file));
    }

    emitDeleted(file) {
      this.dispatchEvent(new HMREvent(HMR.DELETE, file));
    }

    emitRenamed(file, oldFile) {
      this.dispatchEvent(new HMREvent(HMR.RENAME, file, oldFile));
    }
  }

  const instance = new HMR();

  if (typeof window !== "undefined") {
    window.HMR = instance;
  }

  /**
   * @param {MessageEvent<string>} ev
   */
  function onHmrEvent(ev) {
    if (ev.data.startsWith("changed:")) {
      const changedFile = ev.data.substring(8);
      instance.emitChanged(changedFile);
    } else if (ev.data.startsWith("created:")) {
      instance.emitCreated(ev.data.substring(8));
    } else if (ev.data.startsWith("deleted:")) {
      instance.emitDeleted(ev.data.substring(8));
    } else if (ev.data.startsWith("renamed:")) {
      const [file, oldFile] = ev.data.substring(8).split(":");
      instance.emitRenamed(file, oldFile);
    }
  }

  const socket = new WebSocket("ws://" + window.location.host + "/__serve_hmr");
  socket.onmessage = onHmrEvent;
  console.log("HMR enabled");
})();
