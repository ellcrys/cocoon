'use strict';
 
var gulp = require('gulp');
var sass = require('gulp-sass');
var rename = require('gulp-rename');
 
gulp.task('sass', function () {
  return gulp.src('./src/styles/importer.scss')
    .pipe(sass().on('error', sass.logError))
    .pipe(rename('canvas.css'))
    .pipe(gulp.dest('./src/dist'));
});
 
gulp.task('sass:watch', function () {
  gulp.watch('./src/styles/**/*.scss', ['sass']);
});