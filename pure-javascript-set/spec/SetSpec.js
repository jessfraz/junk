describe("Set", function() {
    var emptySet;
    var singletonSet;
    var multiSet;

    beforeEach(function() {
        emptySet = new Set();
        singletonSet = new Set();
        multiSet = new Set();
    });

    it("emptySet should be empty", function() {
        expect(emptySet.isEmpty()).toBe(true);
    });

    describe("when items have been added", function() {
        beforeEach(function() {
            singletonSet.add(1);
            multiSet.add(2);
            multiSet.add(3);
        });

        it("singletonSet should not be empty", function() {
            expect(singletonSet.isEmpty()).toBe(false);
        });

        it("multiSet should not be empty", function() {
            expect(multiSet.isEmpty()).toBe(false);
        });

        it("should return if it contains", function() {
            expect(singletonSet.contains(2)).toBe(false);
            expect(multiSet.contains(2)).toBe(true);
            expect(multiSet.contains(3)).toBe(true);
        });

        it("should return correct size", function() {
            expect(singletonSet.contains(2)).toBe(false);
            expect(multiSet.size).toEqual(2);
            multiSet.add(4);
            expect(multiSet.size).toEqual(3);
        });
    });

    describe("when items have been removed", function() {
        beforeEach(function() {
            singletonSet.add(1);
            multiSet.add(2);
            multiSet.add(3);
            multiSet.add(6);
        });

        it("singletonSet should be empty", function() {
            singletonSet.remove(1);
            expect(singletonSet.isEmpty()).toBe(true);
        });

        it("multiSet should not contain and size decrease", function() {
            multiSet.remove(3);
            expect(multiSet.contains(3)).toBe(false);
            expect(multiSet.size).toEqual(2);
        });

        it("should throw an exception if set doesn't contain item to remove", function() {
            expect(function() {
                multiSet.remove(7);
            }).toThrow("set does not contain item");

            expect(function() {
                emptySet.remove(1);
            }).toThrow("set does not contain item");
        });

    });
});
