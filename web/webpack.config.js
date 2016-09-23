var path = require('path');
var webpack = require('webpack');
var HtmlWebpackPlugin = require('html-webpack-plugin')
var CleanWebpackPlugin = require('clean-webpack-plugin')
var Promise = require('es6-promise').Promise

module.exports = {
	devtool: 'eval',
	entry: {
		app: './src/app.jsx',
		api: './src/api.jsx',
		front: './src/front.jsx',
		ingame: './src/ingame.jsx',
		queue: './src/queue.jsx',
		rank: './src/rank.jsx'
	},
	output: {
		path: path.join(__dirname, 'dist/js'),
		filename: '[name].js',
		publicPath: '/js/'
	},
	plugins: [
		new webpack.HotModuleReplacementPlugin(),
		new HtmlWebpackPlugin({
			template: 'src/assets/index.ejs',
			inject: false,
			filename: '../index.html'
		}),
		new HtmlWebpackPlugin({
			template: 'src/assets/api.ejs',
			inject: false,
			filename: '../api.html'
		}),
		new HtmlWebpackPlugin({
			template: 'src/assets/front.ejs',
			inject: false,
			filename: '../front.html'
		}),
		new HtmlWebpackPlugin({
			template: 'src/assets/ingame.ejs',
			inject: false,
			filename: '../ingame.html'
		}),
		new HtmlWebpackPlugin({
			template: 'src/assets/queue.ejs',
			inject: false,
			filename: '../queue.html'
		}),
		new HtmlWebpackPlugin({
			template: 'src/assets/rank.ejs',
			inject: false,
			filename: '../rank.html'
		}),
		new CleanWebpackPlugin(['dist'])
	],
	module: {
		loaders: [{
			test: /\.jsx$/,
			loaders: ['react-hot', 'babel'],
			include: path.join(__dirname, 'src')
		}, {
			test: /\.css$/,
			loaders: [
				'style?sourceMap',
				'css?modules&importLoaders=1&localIdentName=[path]___[name]__[local]___[hash:base64:5]'
			],
			include: path.join(__dirname, 'src/styles')
		}, {
			test: /\.(png|jpg|gif)$/,
			loader: 'file-loader?name=../img/img-[hash:6].[ext]'
		}, {
			test: /\.(ttf|otf|eot|svg|woff(2)?)(\?[a-z0-9]+)?$/,
			loader: 'file-loader?name=../font/[name].[ext]'
		}]
	}
};
