// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "forge-std/Test.sol";
import {SimpleToken} from "../src/SimpleToken.sol";

contract SimpleTokenTest is Test {
    SimpleToken public token;
    address alice = address(0x1);
    address bob = address(0x2);

    function setUp() public {
        token = new SimpleToken(1000 ether);
        token.mint(alice, 500 ether);
    }

    function testTransfer() public {
        vm.prank(alice);
        token.transfer(bob, 100 ether);
        assertEq(token.balanceOf(bob), 100 ether);
        assertEq(token.balanceOf(alice), 400 ether);
    }

    function testApprove() public {
        vm.prank(alice);
        token.approve(bob, 200 ether);
        assertEq(token.allowance(alice, bob), 200 ether);
    }

    function testTransferFrom() public {
        vm.prank(alice);
        token.approve(bob, 100 ether);
        vm.prank(bob);
        token.transferFrom(alice, bob, 50 ether);
        assertEq(token.balanceOf(bob), 50 ether);
        assertEq(token.allowance(alice, bob), 50 ether);
    }
}
