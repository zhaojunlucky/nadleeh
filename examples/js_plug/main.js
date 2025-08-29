var m = require('./core.mjs')
m.test()
console.log(`hello from plugin: ${env.get('data')}`)
console.log(JSON.stringify(sys))