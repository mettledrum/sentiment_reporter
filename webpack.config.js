var webpack = require('webpack');
var path = require('path');

var BUILD_DIR = path.resolve(__dirname, 'client/public');
var ES6_DIR = path.resolve(__dirname, 'client/es6');

var config = {
    module : {
        loaders : [
            {
                test : /\.jsx?/,
                include : ES6_DIR,
                loader : 'babel'
            }
        ]
    },
    entry: ES6_DIR + '/app.jsx',
    output: {
        path: BUILD_DIR,
        filename: 'bundle.js'
    }
};

module.exports = config;
