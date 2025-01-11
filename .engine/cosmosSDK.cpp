#include <cosmwasm.hpp>
#include <vector>
#include <string>

// Token denomination constants
namespace denom {
    const std::string ATOM = "uatom";            // Cosmos Hub token
    const std::string IBC_BTC = "ibc/BTC_HASH"; // IBC Bitcoin token
    const std::string IBC_ETH = "ibc/ETH_HASH"; // IBC Ethereum token
    const std::string OSMO = "ibc/OSMO_HASH";   // IBC Osmosis token
    const std::string JUNO = "ibc/JUNO_HASH";   // IBC Juno token
    const std::string STARS = "ibc/STARS_HASH"; // IBC Stargaze token
}

// Recipient structure
struct Recipient {
    std::string address;  // Recipient's wallet address
    uint64_t amount;      // Amount to send (in microATOM or token denomination)
};

// Input message structure
struct ReimburseMsg {
    std::vector<Recipient> recipients; // List of recipients with amounts
};

// Main contract logic
class Contract : public cosmwasm::Contract<Contract> {
public:
    cosmwasm::Response execute(const cosmwasm::MessageInfo& info, const ReimburseMsg& msg) {
        // Validate provided funds
        if (info.funds.empty()) {
            throw std::runtime_error("No funds provided");
        }

        uint64_t total_required = calculate_total_amount(msg.recipients);
        uint64_t provided_funds = info.funds.at(0).amount;

        if (provided_funds < total_required) {
            throw std::runtime_error("Insufficient funds provided");
        }

        // Generate BankMsg::Send messages for recipients
        auto bank_msgs = create_bank_msgs(msg.recipients);

        // Create response with messages and attributes
        return cosmwasm::Response()
            .add_messages(bank_msgs)
            .add_attribute("action", "reimburse")
            .add_attribute("sender", info.sender);
    }

private:
    // Helper function to calculate the total amount needed
    uint64_t calculate_total_amount(const std::vector<Recipient>& recipients) {
        uint64_t total = 0;
        for (const auto& recipient : recipients) {
            total += recipient.amount;
        }
        return total;
    }

    // Helper function to create BankMsg::Send messages
    std::vector<cosmwasm::BankMsg> create_bank_msgs(const std::vector<Recipient>& recipients) {
        std::vector<cosmwasm::BankMsg> messages;
        for (const auto& recipient : recipients) {
            messages.emplace_back(cosmwasm::BankMsg::Send{
                recipient.address,
                {{ denom::ATOM, recipient.amount }} // Currently defaulting to ATOM
            });
        }
        return messages;
    }
};

