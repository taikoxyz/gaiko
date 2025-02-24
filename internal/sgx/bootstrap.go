package sgx

import "github.com/urfave/cli/v2"

// pub fn bootstrap(global_opts: GlobalOpts) -> Result<()> {
//     // Generate a new key pair
//     let key_pair = generate_key();
//     // Store it on disk encrypted inside SGX so we can reuse it between program runs
//     let privkey_path = global_opts.secrets_dir.join(PRIV_KEY_FILENAME);
//     save_priv_key(&key_pair, &privkey_path)?;
//     // Get the public key from the pair
//     println!("Public key: 0x{}", key_pair.public_key());
//     let new_instance = public_key_to_address(&key_pair.public_key());
//     println!("Instance address: {new_instance}");
//     // Store the attestation with the new public key
//     save_attestation_user_report_data(new_instance)?;
//     // Store all this data for future use on disk (no encryption necessary)
//     let quote = get_sgx_quote()?;
//     let bootstrap_details_file_path = global_opts.config_dir.join(BOOTSTRAP_INFO_FILENAME);
//     save_bootstrap_details(&key_pair, new_instance, quote, &bootstrap_details_file_path)?;
//     println!(
//         "Bootstrap details saved in {}",
//         bootstrap_details_file_path.display()
//     );
//     println!("Encrypted private key saved in {}", privkey_path.display());
//     Ok(())
// }

func BootStrap(ctx *cli.Context) error {
	panic("todo")
}
