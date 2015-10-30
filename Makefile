build-debs:
	./build-deb $(repo)

build-rpms:
	./build-rpm $(repo)

release-debs:
	./release-deb $(pkg)

release-rpms:
	./release-rpm $(pkg)

clean:
	rm -rf .build
	rm -rf .docker
	rm -rf bundles
