pragma solidity ^0.5.15;

contract Test {
    constructor() public {
        require(1 == 0, "test revert");
    }
}
