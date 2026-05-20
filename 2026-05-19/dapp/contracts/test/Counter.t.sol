// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "forge-std/Test.sol";
import {Counter} from "../src/Counter.sol";

contract CounterTest is Test {
    Counter public counter;

    function setUp() public {
        counter = new Counter();
    }

    function test_InitialCount() public view {
        assertEq(counter.count(), 0);
    }

    function test_Increment() public {
        counter.increment();
        assertEq(counter.count(), 1);
        counter.increment();
        counter.increment();
        assertEq(counter.count(), 3);
    }

    function test_Decrement() public {
        counter.increment();
        counter.increment();
        counter.decrement();
        assertEq(counter.count(), 1);
    }

    function test_RevertWhen_DecrementZero() public {
        vm.expectRevert("Count cannot be negative");
        counter.decrement();
    }

    function test_SetCount_OnlyOwner() public {
        counter.setCount(42);
        assertEq(counter.count(), 42);
    }

    function test_RevertWhen_NonOwnerSetsCount() public {
        vm.prank(address(0x123));
        vm.expectRevert("Caller is not owner");
        counter.setCount(99);
    }

    function test_TransferOwnership() public {
        address newOwner = address(0x456);
        counter.transferOwnership(newOwner);
        assertEq(counter.owner(), newOwner);

        vm.prank(newOwner);
        counter.setCount(100);
        assertEq(counter.count(), 100);
    }
}
