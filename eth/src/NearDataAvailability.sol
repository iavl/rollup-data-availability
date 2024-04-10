// SPDX-License-Identifier: AGPL-3.0
pragma solidity >=0.8.25;

import { IDataAvailabilityProtocol } from "@polygon/zkevm-contracts/interfaces/IDataAvailabilityProtocol.sol";
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";

/**
 * @dev Struct to store the data availability batch, transaction verification on ethereum and transaction submission on
 * NEAR
 *
 */
struct VerifiedBatch {
    bytes32 id;
    bytes32 verifyTxHash;
    bytes32 submitTxId;
}

/*
 * Contract responsible for storing the lookup information for the status of each NEARDA batch
 * It is heavily modeled after the requirements from polygon CDK
 */
contract NearDataAvailability is Initializable, IDataAvailabilityProtocol, OwnableUpgradeable {
    // Name of the data availability protocol
    string internal constant _PROTOCOL_NAME = "NearProtocol";

    // The amount of batches that we track is available, one NEAR epoch.
    // note, they are still available via archival, with additional flows.
    uint256 public constant _STORED_BATCH_AMT = 12;

    // The batches that have been made available, keyed by bucket id
    // and dusts the ones more than _STORED_BATCH_AMT
    VerifiedBatch[_STORED_BATCH_AMT] public batchInfo;

    uint256 public _nextBucketId;

    /**
     * @dev Emitted when the DA batch is made available, used to determine if the batch has been proven
     * @param batch Batch of data that has been made available
     * @param bucket id of the batch
     */
    event IsAvailable(uint256 bucket, VerifiedBatch batch);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address initialOwner) public initializer {
        __Ownable_init(initialOwner);
    }

    /**
     * @notice Verifies that the given signedHash has been signed by requiredAmountOfSignatures committee members
     * @param dataAvailabilityBatch blarg
     */
    function verifyMessage(bytes32, /*hash*/ bytes calldata dataAvailabilityBatch) external view {
        VerifiedBatch storage item;
        bytes32 batchId = abi.decode(dataAvailabilityBatch, (bytes32));
        for (uint256 i = 0; i < batchInfo.length; i++) {
            item = batchInfo[i];
            if (item.id == batchId) {
                return;
            }
        }
        revert("Batch not found");
    }

    // TODO: ensurerole
    // TODO: test me
    function notifyAvailable(VerifiedBatch memory verifiedBatch) external {
        uint256 bucket = _nextBucketId;

        VerifiedBatch storage b = batchInfo[bucket];
        b.id = verifiedBatch.id;
        b.verifyTxHash = verifiedBatch.verifyTxHash;
        b.submitTxId = verifiedBatch.submitTxId;

        emit IsAvailable(bucket, verifiedBatch);
        _nextBucketId = (bucket + 1) % _STORED_BATCH_AMT;
    }

    /// Registers a listener to the DA light client
    function registerListener(address publisher) public returns (bool) {
        // Make a call to the publisher
    }

    /**
     * @notice Return the protocol name
     */
    function getProcotolName() external pure override returns (string memory) {
        return _PROTOCOL_NAME;
    }
}
