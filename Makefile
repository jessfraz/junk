build-debs:
	./build-deb $(repo)

build-rpms:
	./build-rpm $(repo)

release-debs:
	./release-deb $(pkg) $(component)

release-rpms:
	./release-rpm $(pkg) $(component)

clean:
	rm -rf .build
	rm -rf .docker
	rm -rf bundles
