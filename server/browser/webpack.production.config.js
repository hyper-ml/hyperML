
var webpack = require('webpack')
var path = require('path')
var PurgecssPlugin = require('purgecss-webpack-plugin')
var glob = require('glob-all')

var exports = {
    context: __dirname,
    mode: 'production',
    entry: [
        path.resolve(__dirname, 'src/index.js')
    ],
    output: {
        path: path.resolve(__dirname, 'production'),
        filename: 'index.bundle.js'
    },
    module: {
        rules: [{
            test: /\.js$/,
            use: [{
                loader: 'babel-loader',
                options: {
                    presets: ['react', 'es2015']
                },
            }],
            exclude: /(node_modules)/
        },{
            test: /\.s[ac]ss$/i,
            use: [
            // Creates `style` nodes from JS strings
            'style-loader',
            // Translates CSS into CommonJS
            'css-loader',
            // Compiles Sass to CSS
            'sass-loader',
            ],
        }, {
            test: /\.css$/,
            use: [{
              loader: 'style-loader'
            }, {
              loader: 'css-loader'
            }]
          }, {
            test: /\.(eot|woff|woff2|ttf|svg|png)/,
            use: [{
              loader: 'url-loader'
            }]
          }
        ]
    },
    node: {
        fs:'empty'
    },
    plugins: [
        new webpack.ContextReplacementPlugin(/moment[\\\/]locale$/, /^\.\/(en)$/),
        new PurgecssPlugin({
            paths: glob.sync([
              path.join(__dirname, 'app/index.html'),
              path.join(__dirname, 'app/js/*.js')
            ])
        })
    ]

}

if (process.env.NODE_ENV === 'dev') {
    exports.entry = [
      'webpack-dev-server/client?http://localhost:8080',
      path.resolve(__dirname, 'app/index.js')
    ]
}

module.exports = exports
  