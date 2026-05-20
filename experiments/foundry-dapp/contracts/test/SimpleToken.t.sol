// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "forge-std/Test.sol";
import {SimpleToken} from "../src/SimpleToken.sol";

contract SimpleTokenTest is Test {
    SimpleToken public token;
    address public alice = address(0x1234);
    address public bob = address(0x5678);
    address public attacker = address(0xDEAD);

    function setUp() public {
        token = new SimpleToken(1000);
        // 给 alice 100 个 token
        vm.prank(address(this));
        token.transfer(alice, 100 * 10 ** 18);
    }

    function test_InitialSupply() public view {
        assertEq(token.totalSupply(), 1000 * 10 ** 18);
        assertEq(token.balanceOf(address(this)), 900 * 10 ** 18);
    }

    function test_Transfer() public {
        vm.prank(alice);
        bool ok = token.transfer(bob, 50 * 10 ** 18);
        assertTrue(ok);
        assertEq(token.balanceOf(alice), 50 * 10 ** 18);
        assertEq(token.balanceOf(bob), 50 * 10 ** 18);
    }

    function test_RevertWhen_InsufficientBalance() public {
        vm.prank(alice);
        vm.expectRevert("Insufficient balance");
        token.transfer(bob, 200 * 10 ** 18);
    }

    function test_Approve() public {
        vm.prank(alice);
        token.approve(bob, type(uint256).max);
        assertEq(token.allowance(alice, bob), type(uint256).max);
    }

    function test_TransferFrom() public {
        vm.prank(alice);
        token.approve(bob, 30 * 10 ** 18);

        vm.prank(bob);
        token.transferFrom(alice, attacker, 30 * 10 ** 18);

        assertEq(token.balanceOf(alice), 70 * 10 ** 18);
        assertEq(token.balanceOf(attacker), 30 * 10 ** 18);
        assertEq(token.allowance(alice, bob), 0);
    }

    function test_RevertWhen_ExceedsAllowance() public {
        vm.prank(alice);
        token.approve(bob, 10 * 10 ** 18);

        vm.prank(bob);
        vm.expectRevert("Insufficient allowance");
        token.transferFrom(alice, attacker, 20 * 10 ** 18);
    }
}
