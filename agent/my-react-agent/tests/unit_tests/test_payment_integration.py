"""Tests for MiniPay client and payment tools integration."""

import pytest
from unittest.mock import AsyncMock, patch, MagicMock
import json

from react_agent.minipay_client import (
    MiniPayClient,
    ChargeResponse,
    RefundResponse,
    BalanceResponse,
    MetricsResponse,
)
from react_agent.payment_tools import (
    create_charge,
    refund_charge,
    get_account_balance,
    get_system_metrics,
    check_backend_health,
)


@pytest.mark.asyncio
async def test_minipay_client_create_charge():
    """Test MiniPayClient.create_charge method."""
    mock_response = MagicMock()
    mock_response.json.return_value = {
        "id": "txn_test123",
        "amount": 1000,
        "currency": "usd",
        "customer": "cust_test",
        "status": "succeeded",
        "idempotency_key": "idem_test",
        "created_at": "2025-11-03T10:00:00Z",
    }
    mock_response.raise_for_status = MagicMock()

    with patch("react_agent.minipay_client.httpx.AsyncClient") as mock_client:
        mock_instance = AsyncMock()
        mock_client.return_value = mock_instance
        mock_instance.post.return_value = mock_response

        client = MiniPayClient()
        result = await client.create_charge(
            amount=1000, currency="usd", customer="cust_test"
        )

        assert result.id == "txn_test123"
        assert result.amount == 1000
        assert result.status == "succeeded"


@pytest.mark.asyncio
async def test_minipay_client_get_balance():
    """Test MiniPayClient.get_balance method."""
    mock_response = MagicMock()
    mock_response.json.return_value = {
        "successful_transactions": 10,
        "refunded_transactions": 2,
        "balance": 5000,
    }
    mock_response.raise_for_status = MagicMock()

    with patch("react_agent.minipay_client.httpx.AsyncClient") as mock_client:
        mock_instance = AsyncMock()
        mock_client.return_value = mock_instance
        mock_instance.get.return_value = mock_response

        client = MiniPayClient()
        result = await client.get_balance()

        assert result.balance == 5000
        assert result.successful_transactions == 10


@pytest.mark.asyncio
async def test_payment_tool_create_charge():
    """Test create_charge payment tool."""
    mock_charge_response = ChargeResponse(
        id="txn_test123",
        amount=1000,
        currency="usd",
        customer="cust_test",
        status="succeeded",
        idempotency_key="idem_test",
        created_at="2025-11-03T10:00:00Z",
    )

    with patch(
        "react_agent.payment_tools.MiniPayClient"
    ) as mock_client_class:
        mock_instance = AsyncMock()
        mock_client_class.return_value.__aenter__.return_value = mock_instance
        mock_instance.create_charge.return_value = mock_charge_response

        result = await create_charge(amount=1000, customer="cust_test")
        result_dict = json.loads(result)

        assert result_dict["success"] is True
        assert result_dict["transaction_id"] == "txn_test123"
        assert result_dict["amount"] == 1000


@pytest.mark.asyncio
async def test_payment_tool_get_balance():
    """Test get_account_balance payment tool."""
    mock_balance_response = BalanceResponse(
        successful_transactions=10,
        refunded_transactions=2,
        balance=5000,
    )

    with patch(
        "react_agent.payment_tools.MiniPayClient"
    ) as mock_client_class:
        mock_instance = AsyncMock()
        mock_client_class.return_value.__aenter__.return_value = mock_instance
        mock_instance.get_balance.return_value = mock_balance_response

        result = await get_account_balance()
        result_dict = json.loads(result)

        assert result_dict["success"] is True
        assert result_dict["balance"] == 5000


@pytest.mark.asyncio
async def test_payment_tool_create_charge_error():
    """Test create_charge tool error handling."""
    with patch(
        "react_agent.payment_tools.MiniPayClient"
    ) as mock_client_class:
        mock_instance = AsyncMock()
        mock_client_class.return_value.__aenter__.return_value = mock_instance
        mock_instance.create_charge.side_effect = Exception("Connection failed")

        result = await create_charge(amount=1000, customer="cust_test")
        result_dict = json.loads(result)

        assert result_dict["success"] is False
        assert "error" in result_dict


@pytest.mark.asyncio
async def test_payment_tool_refund_charge():
    """Test refund_charge payment tool."""
    mock_refund_response = RefundResponse(
        id="txn_test123",
        amount=1000,
        status="refunded",
        refunded_at="2025-11-03T10:05:00Z",
    )

    with patch(
        "react_agent.payment_tools.MiniPayClient"
    ) as mock_client_class:
        mock_instance = AsyncMock()
        mock_client_class.return_value.__aenter__.return_value = mock_instance
        mock_instance.refund_transaction.return_value = mock_refund_response

        result = await refund_charge("txn_test123")
        result_dict = json.loads(result)

        assert result_dict["success"] is True
        assert result_dict["status"] == "refunded"


@pytest.mark.asyncio
async def test_payment_tool_get_metrics():
    """Test get_system_metrics payment tool."""
    mock_metrics_response = MetricsResponse(
        total_charges=100,
        total_refunds=5,
        pending_webhooks=3,
        delivered_webhooks=95,
        failed_webhooks=2,
        webhook_retries=12,
    )

    with patch(
        "react_agent.payment_tools.MiniPayClient"
    ) as mock_client_class:
        mock_instance = AsyncMock()
        mock_client_class.return_value.__aenter__.return_value = mock_instance
        mock_instance.get_metrics.return_value = mock_metrics_response

        result = await get_system_metrics()
        result_dict = json.loads(result)

        assert result_dict["success"] is True
        assert result_dict["total_charges"] == 100
        assert result_dict["pending_webhooks"] == 3


@pytest.mark.asyncio
async def test_payment_tool_check_health():
    """Test check_backend_health payment tool."""
    with patch(
        "react_agent.payment_tools.MiniPayClient"
    ) as mock_client_class:
        mock_instance = AsyncMock()
        mock_client_class.return_value.__aenter__.return_value = mock_instance
        mock_instance.health_check.return_value = True

        result = await check_backend_health()
        result_dict = json.loads(result)

        assert result_dict["success"] is True
        assert result_dict["healthy"] is True


@pytest.mark.asyncio
async def test_minipay_client_health_check_failure():
    """Test MiniPayClient.health_check with backend down."""
    with patch("react_agent.minipay_client.httpx.AsyncClient") as mock_client:
        mock_instance = AsyncMock()
        mock_client.return_value = mock_instance
        mock_instance.get.side_effect = Exception("Connection refused")

        client = MiniPayClient()
        result = await client.health_check()

        assert result is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
