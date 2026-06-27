;(function () {
  'use strict'

  const _nativeToString = Function.prototype.toString
  const _fakeNatives = new WeakMap()
  const _nativeToStringFn = function toString() {
    if (_fakeNatives.has(this)) return _fakeNatives.get(this)
    return _nativeToString.call(this)
  }
  _fakeNatives.set(_nativeToStringFn, 'function toString() { [native code] }')
  Object.defineProperty(Function.prototype, 'toString', {
    value: _nativeToStringFn, writable: true, configurable: true,
  })

  function makeNative(fn, name) {
    _fakeNatives.set(fn, `function ${name || fn.name}() { [native code] }`)
    return fn
  }

  if (!window.chrome) {
    window.chrome = {
      app: {
        isInstalled: false,
        InstallState: { DISABLED: 'disabled', INSTALLED: 'installed', NOT_INSTALLED: 'not_installed' },
        RunningState: { CANNOT_RUN: 'cannot_run', READY_TO_RUN: 'ready_to_run', RUNNING: 'running' },
      },
      runtime: {
        OnInstalledReason: { CHROME_UPDATE: 'chrome_update', INSTALL: 'install', SHARED_MODULE_UPDATE: 'shared_module_update', UPDATE: 'update' },
        OnRestartRequiredReason: { APP_UPDATE: 'app_update', OS_UPDATE: 'os_update', PERIODIC: 'periodic' },
        PlatformArch: { ARM: 'arm', ARM64: 'arm64', MIPS: 'mips', MIPS64: 'mips64', X86_32: 'x86-32', X86_64: 'x86-64' },
        PlatformOs: { ANDROID: 'android', CROS: 'cros', LINUX: 'linux', MAC: 'mac', OPENBSD: 'openbsd', WIN: 'win' },
        RequestUpdateCheckStatus: { NO_UPDATE: 'no_update', THROTTLED: 'throttled', UPDATE_AVAILABLE: 'update_available' },
        id: undefined,
        connect: makeNative(function connect() {}, 'connect'),
        sendMessage: makeNative(function sendMessage() {}, 'sendMessage'),
      },
      loadTimes: makeNative(function loadTimes() { return {} }, 'loadTimes'),
      csi: makeNative(function csi() { return {} }, 'csi'),
    }
  }

  if (navigator.plugins.length === 0) {
    Object.defineProperty(navigator, 'plugins', {
      get: makeNative(function () {
        const arr = [
          { name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format' },
          { name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: '' },
          { name: 'Native Client', filename: 'internal-nacl-plugin', description: '' },
        ]
        arr.__proto__ = PluginArray.prototype
        return arr
      }, 'get plugins'),
      configurable: true,
    })
  }

  if (!navigator.languages || navigator.languages.length === 0) {
    Object.defineProperty(navigator, 'languages', {
      get: makeNative(function () { return ['en-US', 'en'] }, 'get languages'),
      configurable: true,
    })
  }

  if (window.Notification && Notification.permission === 'denied') {
    Object.defineProperty(Notification, 'permission', {
      get: makeNative(function () { return 'default' }, 'get permission'),
      configurable: true,
    })
  }

  if (navigator.permissions) {
    const _origQuery = navigator.permissions.query.bind(navigator.permissions)
    navigator.permissions.__proto__.query = makeNative(function query(parameters) {
      if (parameters && parameters.name === 'notifications') {
        return Promise.resolve({ state: window.Notification ? Notification.permission : 'default' })
      }
      return _origQuery(parameters)
    }, 'query')
  }

  Object.defineProperty(navigator, 'maxTouchPoints', {
    get: makeNative(function () { return 1 }, 'get maxTouchPoints'),
    configurable: true,
  })

  Object.defineProperty(navigator, 'hardwareConcurrency', {
    get: makeNative(function () { return 8 }, 'get hardwareConcurrency'), configurable: true,
  })
  if ('deviceMemory' in navigator) {
    Object.defineProperty(navigator, 'deviceMemory', {
      get: makeNative(function () { return 8 }, 'get deviceMemory'), configurable: true,
    })
  }

  if (navigator.connection) {
    Object.defineProperty(navigator.connection, 'rtt', {
      get: makeNative(function () { return 100 }, 'get rtt'), configurable: true,
    })
  }

  const W = window.innerWidth  || 1920
  const H = window.innerHeight || 1080
  const screenProps = { width: W, height: H, availWidth: W, availHeight: H - 40,
                        availLeft: 0, availTop: 0, colorDepth: 24, pixelDepth: 24 }
  for (const [k, v] of Object.entries(screenProps)) {
    const val = v
    Object.defineProperty(screen, k, {
      get: makeNative(function () { return val }, `get ${k}`), configurable: true,
    })
  }

  if (!navigator.mediaDevices) {
    Object.defineProperty(navigator, 'mediaDevices', {
      get: makeNative(function () {
        return { enumerateDevices: () => Promise.resolve([]) }
      }, 'get mediaDevices'),
      configurable: true,
    })
  }

  const _origToDataURL = HTMLCanvasElement.prototype.toDataURL
  HTMLCanvasElement.prototype.toDataURL = makeNative(function (type) {
    const ctx = this.getContext('2d')
    if (ctx) {
      const d = ctx.getImageData(0, 0, this.width, this.height)
      d.data[0] ^= 1
      ctx.putImageData(d, 0, 0)
    }
    return _origToDataURL.apply(this, arguments)
  }, 'toDataURL')

  const _origGetImageData = CanvasRenderingContext2D.prototype.getImageData
  CanvasRenderingContext2D.prototype.getImageData = makeNative(function (x, y, w, h) {
    const d = _origGetImageData.call(this, x, y, w, h)
    d.data[0] ^= 1
    return d
  }, 'getImageData')

  const _origGetParam = WebGLRenderingContext.prototype.getParameter
  WebGLRenderingContext.prototype.getParameter = makeNative(function (param) {
    if (param === 37445) return 'Intel Inc.'
    if (param === 37446) return 'Intel Iris OpenGL Engine'
    return _origGetParam.call(this, param)
  }, 'getParameter')

  if (window.WebGL2RenderingContext) {
    const _orig2 = WebGL2RenderingContext.prototype.getParameter
    WebGL2RenderingContext.prototype.getParameter = makeNative(function (param) {
      if (param === 37445) return 'Intel Inc.'
      if (param === 37446) return 'Intel Iris OpenGL Engine'
      return _orig2.call(this, param)
    }, 'getParameter')
  }

  const AC = window.AudioContext || window.webkitAudioContext
  if (AC) {
    const _origOsc = AC.prototype.createOscillator
    AC.prototype.createOscillator = makeNative(function () {
      const osc = _origOsc.call(this)
      const _origSetVal = osc.frequency.setValueAtTime.bind(osc.frequency)
      osc.frequency.setValueAtTime = (v, t) => _origSetVal(v + 0.0001, t)
      return osc
    }, 'createOscillator')
  }
})()
