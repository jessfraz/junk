module.exports = function(grunt) {
	require('matchdep').filterDev('grunt-*').forEach(grunt.loadNpmTasks);

	grunt.initConfig({
		pkg: grunt.file.readJSON('package.json'),
		banner: '/*\n<%= pkg.name %> - v<%= pkg.version %> - ' + '<%= grunt.template.today("yyyy-mm-dd") %>\n<%= pkg.description %>\nLovingly coded by <%= pkg.author.name %>  - <%= pkg.author.url %> \n*/\n',	
		watch: {
            options: {
				livereload: true
			},
			html: {
				files: '**/*.js'
			}
		}
	});
		
	grunt.registerTask('server', [
		'watch'
	]);

	grunt.registerTask('default', 'server');
};
